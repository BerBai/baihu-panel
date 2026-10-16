package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/engigu/baihu-panel/internal/constant"
	"github.com/engigu/baihu-panel/internal/eventbus"
	"github.com/engigu/baihu-panel/internal/logger"
	"github.com/engigu/baihu-panel/internal/middleware"
	"github.com/engigu/baihu-panel/internal/utils"
	"golang.org/x/oauth2"

	"github.com/gin-gonic/gin"
)



// OidcTestRequest 连通性测试请求参数
type OidcTestRequest struct {
	Type        string `json:"type"`
	Issuer      string `json:"issuer"`
	AuthURL     string `json:"oauth2_auth_url"`
	TokenURL    string `json:"oauth2_token_url"`
	UserinfoURL string `json:"oauth2_userinfo_url"`
}

// OidcTest 供设置界面测试单点登录提供方网络连通性与发现文档
func (ac *AuthController) OidcTest(c *gin.Context) {
	var req OidcTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数格式错误")
		return
	}

	httpClient := newSSOHTTPClient()

	if req.Type == constant.SSOTypeOAuth2 || req.AuthURL != "" {
		endpoints := map[string]string{
			"授权端点": req.AuthURL,
			"令牌端点": req.TokenURL,
			"用户信息": req.UserinfoURL,
		}
		for name, url := range endpoints {
			u := strings.TrimSpace(url)
			if u == "" {
				utils.BadRequest(c, fmt.Sprintf("%s 不能为空", name))
				return
			}
			reqCtx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			reqObj, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u, nil)
			if err != nil {
				cancel()
				utils.ServerError(c, fmt.Sprintf("%s 地址非法: %v", name, err))
				return
			}
			reqObj.Header.Set("User-Agent", "Baihu-Panel-SSO/1.0")
			resp, err := httpClient.Do(reqObj)
			cancel()
			if err != nil {
				utils.ServerError(c, fmt.Sprintf("%s 无法连接: %v", name, err))
				return
			}
			resp.Body.Close()
		}
		utils.Success(c, gin.H{
			"message": "OAuth 2.0 端点网络连通性测试通过",
		})
		return
	}

	// OIDC 模式
	issuer := NormalizeIssuer(req.Issuer)
	if issuer == "" {
		utils.BadRequest(c, "Issuer（发现文档地址）不能为空")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	ctx = oidc.ClientContext(ctx, httpClient)
	ctx = oidc.InsecureIssuerURLContext(ctx, issuer)

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		altIssuer := strings.TrimRight(issuer, "/")
		if altIssuer == issuer {
			altIssuer = issuer + "/"
		}
		altCtx := oidc.InsecureIssuerURLContext(ctx, altIssuer)
		if p2, err2 := oidc.NewProvider(altCtx, altIssuer); err2 == nil {
			provider = p2
			err = nil
		}
	}
	if err != nil {
		errDetail := err.Error()
		if strings.Contains(errDetail, "404") && strings.Contains(errDetail, "API endpoint not found") {
			utils.ServerError(c, fmt.Sprintf("连接提供方失败 (404 Not Found)。当前身份系统（如 Pocket ID）端点未响应发现文档。请检查：1. Issuer 只需填写根域名（如 https://id.example.com，不可包含 /.well-known 等后缀）；2. 确保反向代理已将 /.well-known 转发至认证容器；或建议直接切换为【通用 OAuth 2.0】模式手动配置端点。原始错误: %s", errDetail))
			return
		}
		utils.ServerError(c, "连接提供方失败: "+errDetail)
		return
	}

	claims := map[string]interface{}{}
	_ = provider.Claims(&claims)

	utils.Success(c, gin.H{
		"message":   "发现文档解析成功，OIDC 通道连通正常",
		"endpoints": provider.Endpoint(),
		"claims":    claims,
	})
}

// startSSOLogin 发起单点登录跳转流程（统一支持 OIDC 与通用 OAuth 2.0）
func (ac *AuthController) startSSOLogin(c *gin.Context, cfg *oidcConfig) {
	if constant.DemoMode {
		utils.BadRequest(c, "演示模式下禁止使用单点登录")
		return
	}
	if cfg == nil {
		utils.BadRequest(c, "该单点登录通道未启用或配置不完整")
		return
	}

	_, oauthCfg, err := ac.oidcProviderFor(c.Request.Context(), cfg)
	if err != nil {
		logger.Errorf("[SSO] 无法连接单点登录提供方 (Type: %s): %v", cfg.Type, err)
		utils.ServerError(c, "无法连接单点登录提供方: "+err.Error())
		return
	}

	state := utils.RandomString(32)
	postLogin := "/"
	if to := c.Query("redirect"); strings.HasPrefix(to, "/") && !strings.HasPrefix(to, "//") {
		postLogin = to
	}
	verifier := oauth2.GenerateVerifier()

	c.SetSameSite(http.SameSiteLaxMode)
	// Cookie 保存「state.通道类型.回跳路径」，回调时精准识别通道
	c.SetCookie(oidcStateCookie, state+"."+cfg.Type+"."+postLogin, int(oidcStateTTL.Seconds()), "/", "", false, true)
	c.SetCookie(oidcVerifierCookie, verifier, int(oidcStateTTL.Seconds()), "/", "", false, true)

	authURL := oauthCfg.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	c.Redirect(302, authURL)
}

// handleSSOCallback 处理单点登录授权回调（统一支持 OIDC 与通用 OAuth 2.0）
func (ac *AuthController) handleSSOCallback(c *gin.Context, defaultCfg *oidcConfig) {
	if constant.DemoMode {
		utils.BadRequest(c, "演示模式下禁止使用单点登录")
		return
	}

	stateCookie, err := c.Cookie(oidcStateCookie)
	if err != nil || stateCookie == "" {
		utils.Unauthorized(c, "单点登录会话已失效，请重新登录")
		return
	}
	verifier, _ := c.Cookie(oidcVerifierCookie)
	middleware.ClearCookie(c, oidcStateCookie)
	middleware.ClearCookie(c, oidcVerifierCookie)

	// 解析 cookie: state.type.redirect
	parts := strings.SplitN(stateCookie, ".", 3)
	baseState := parts[0]
	channelType := ""
	postLogin := "/"
	if len(parts) == 3 {
		channelType = parts[1]
		postLogin = parts[2]
	} else if len(parts) == 2 {
		postLogin = parts[1]
	}

	if got := c.Query("state"); got == "" || got != baseState {
		utils.Unauthorized(c, "单点登录 state 校验失败")
		return
	}
	if e := c.Query("error"); e != "" {
		utils.Unauthorized(c, "第三方授权失败: "+e)
		return
	}
	code := c.Query("code")
	if code == "" {
		utils.Unauthorized(c, "缺少授权码")
		return
	}

	// 识别目标通道配置
	cfg := defaultCfg
	if channelType == constant.SSOTypeOAuth2 {
		if o2Cfg := ac.getOAuth2Config(c); o2Cfg != nil {
			cfg = o2Cfg
		}
	} else if channelType == constant.SSOTypeOIDC {
		if oidcCfg := ac.getOidcConfig(c); oidcCfg != nil {
			cfg = oidcCfg
		}
	}
	if cfg == nil {
		cfg = ac.oidcConfig(c)
	}
	if cfg == nil {
		utils.BadRequest(c, "单点登录通道未启用或配置不完整")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	verifierObj, oauthCfg, err := ac.oidcProviderFor(ctx, cfg)
	if err != nil {
		logger.Errorf("[SSO] 回调阶段无法连接单点登录提供方: %v", err)
		utils.ServerError(c, "无法连接单点登录提供方: "+err.Error())
		return
	}

	token, err := oauthCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		token, err = oauthCfg.Exchange(ctx, code)
	}
	if err != nil {
		ac.logOidcFailure(c, "", "授权码换取令牌失败: "+err.Error())
		utils.Unauthorized(c, "授权码换取令牌失败")
		return
	}

	var claims oidcClaims
	if cfg.Type == constant.SSOTypeOAuth2 {
		cObj, err := ac.fetchOAuth2Userinfo(ctx, cfg, token.AccessToken)
		if err != nil {
			ac.logOidcFailure(c, "", "获取用户信息失败: "+err.Error())
			utils.Unauthorized(c, "获取用户信息失败: "+err.Error())
			return
		}
		claims = *cObj
	} else {
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok || rawIDToken == "" {
			utils.Unauthorized(c, "OIDC 未返回 id_token")
			return
		}

		idToken, err := verifierObj.Verify(ctx, rawIDToken)
		if err != nil {
			ac.logOidcFailure(c, "", "id_token 校验失败")
			utils.Unauthorized(c, "id_token 校验失败")
			return
		}
		if err := idToken.Claims(&claims); err != nil {
			utils.Unauthorized(c, "无法解析 OIDC 身份信息")
			return
		}
	}

	user, created := ac.resolveOidcUser(cfg, &claims)
	if user == nil {
		ac.logOidcFailure(c, claims.PreferredUsername, "未绑定账号且未开启自动创建")
		utils.Forbidden(c, "当前第三方身份未绑定任何系统账号，请联系管理员")
		return
	}

	expireDays := 7
	if d := utils.ParsePositiveInt(ac.settingsService.Get(constant.SectionSite, constant.KeyCookieDays)); d > 0 {
		expireDays = d
	}
	tokenStr, err := utils.GenerateToken(user.ID, user.Username, user.TokenVersion, expireDays, constant.Secret)
	if err != nil {
		utils.ServerError(c, "登录失败")
		return
	}
	middleware.SetAuthCookie(c, tokenStr, expireDays)

	msg := "单点登录成功 (" + cfg.Type + ")"
	if created {
		msg = "第三方身份首次登录，已自动创建账号 (" + cfg.Type + ")"
	}
	eventbus.DefaultBus.Publish(eventbus.Event{
		Type: constant.EventUserLogin,
		Payload: map[string]interface{}{
			"ip":        c.ClientIP(),
			"username":  user.Username,
			"userAgent": c.GetHeader("User-Agent"),
			"status":    "success",
			"message":   msg,
		},
	})

	if !strings.HasPrefix(postLogin, "/") || strings.HasPrefix(postLogin, "//") {
		postLogin = "/"
	}
	c.Redirect(302, postLogin)
}

// OidcLogin 重定向到 OIDC 授权端点
func (ac *AuthController) OidcLogin(c *gin.Context) {
	ac.startSSOLogin(c, ac.getOidcConfig(c))
}

// OidcCallback 处理 OIDC 授权回调
func (ac *AuthController) OidcCallback(c *gin.Context) {
	ac.handleSSOCallback(c, ac.getOidcConfig(c))
}

// OAuth2Login 重定向到通用 OAuth 2.0 授权端点
func (ac *AuthController) OAuth2Login(c *gin.Context) {
	ac.startSSOLogin(c, ac.getOAuth2Config(c))
}

// OAuth2Callback 处理通用 OAuth 2.0 授权回调
func (ac *AuthController) OAuth2Callback(c *gin.Context) {
	ac.handleSSOCallback(c, ac.getOAuth2Config(c))
}

// OAuth2Test 供前端测试通用 OAuth 2.0 接口连通性
func (ac *AuthController) OAuth2Test(c *gin.Context) {
	ac.OidcTest(c)
}
