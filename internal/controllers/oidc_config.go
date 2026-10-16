package controllers

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/engigu/baihu-panel/internal/constant"
	"github.com/engigu/baihu-panel/internal/services"
	"github.com/engigu/baihu-panel/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

const (
	oidcStateCookie    = "BHOIDCState"
	oidcVerifierCookie = "BHOIDCVerifier"
	oidcStateTTL       = 10 * time.Minute
)

// oidcProviderCache 缓存发现文档与 JWKS，避免每次登录都远程拉取
type oidcProviderCache struct {
	mu       sync.Mutex
	issuer   string
	clientID string
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
}

var oidcCache = &oidcProviderCache{}

type oidcConfig struct {
	Type         string   // "oidc" 或 "oauth2"
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AutoCreate   bool
	DefaultRole  string

	// 通用 OAuth 2.0 专用字段
	AuthURL     string
	TokenURL    string
	UserinfoURL string
	DisplayName string
	UserKey     string
}

// resolveRedirectURL 解析回调地址：优先使用管理员配置；留空时根据当前请求动态推导标准回调地址
func (ac *AuthController) resolveRedirectURL(custom string, defaultPath string, c *gin.Context) string {
	if trimmed := strings.TrimSpace(custom); trimmed != "" {
		return trimmed
	}
	if defaultPath == "" {
		defaultPath = "/api/v1/auth/oidc/callback"
	}
	if c != nil && c.Request != nil {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := c.GetHeader("X-Forwarded-Host")
		if host == "" {
			host = c.Request.Host
		}
		if host != "" {
			return fmt.Sprintf("%s://%s%s", scheme, host, defaultPath)
		}
	}
	return defaultPath
}

// NormalizeIssuer 规范化 Issuer 地址，自动剔除用户误填的 /.well-known/openid-configuration 后缀与多余斜杠
func NormalizeIssuer(raw string) string {
	u := strings.TrimSpace(raw)
	u = strings.TrimRight(u, "/")
	for {
		trimmed := strings.TrimSuffix(u, "/.well-known/openid-configuration")
		trimmed = strings.TrimSuffix(trimmed, "/.well-known/openid-configuration/")
		trimmed = strings.TrimRight(trimmed, "/")
		if trimmed == u {
			break
		}
		u = trimmed
	}
	return u
}

// securityReader 封装对 Security 配置段的高效安全读取、优先级回退与类型解析助手
type securityReader struct {
	ss *services.SettingsService
}

func (r securityReader) get(keys ...string) string {
	if r.ss == nil {
		return ""
	}
	for _, k := range keys {
		if v := strings.TrimSpace(r.ss.Get(constant.SectionSecurity, k)); v != "" {
			return v
		}
	}
	return ""
}

func (r securityReader) getWithDefault(defaultVal string, keys ...string) string {
	if v := r.get(keys...); v != "" {
		return v
	}
	return defaultVal
}

func (r securityReader) bool(keys ...string) bool {
	return r.get(keys...) == "true"
}

func (r securityReader) secret(keys ...string) string {
	raw := r.get(keys...)
	if raw == "" {
		return ""
	}
	return utils.DecryptSecret(raw)
}

func (r securityReader) role(keys ...string) string {
	if r.get(keys...) == constant.AdminRole {
		return constant.AdminRole
	}
	return constant.DefaultRole
}

func (r securityReader) scopes(raw string, defaultScopes ...string) []string {
	seen := make(map[string]bool)
	var list []string
	for _, s := range defaultScopes {
		if s = strings.TrimSpace(s); s != "" && !seen[s] {
			seen[s] = true
			list = append(list, s)
		}
	}
	for _, s := range strings.FieldsFunc(raw, func(c rune) bool {
		return c == ',' || c == ' ' || c == '\n'
	}) {
		if s = strings.TrimSpace(s); s != "" && !seen[s] {
			seen[s] = true
			list = append(list, s)
		}
	}
	return list
}

// ResolveSSOStatus 解析当前单点登录启用状态与展示信息（支持多身份源并行，用于 /public/settings）
func ResolveSSOStatus(settingsService *services.SettingsService) gin.H {
	if constant.DemoMode || settingsService == nil {
		return gin.H{"enabled": false, "providers": []gin.H{}}
	}

	r := securityReader{settingsService}
	var providers []gin.H

	// 1. 检查标准 OIDC 通道
	if r.bool(constant.KeyOidcEnabled) {
		clientID := r.get(constant.KeyOidcClientID)
		issuer := NormalizeIssuer(r.get(constant.KeyOidcIssuer))
		if clientID != "" && issuer != "" {
			providers = append(providers, gin.H{
				"id":        "oidc",
				"type":      constant.SSOTypeOIDC,
				"name":      "标准 OIDC",
				"login_url": "/api/v1/auth/oidc/login",
			})
		}
	}

	// 2. 检查通用 OAuth 2.0 通道（支持独立开关或历史兼容开关）
	oauth2Enabled := r.bool(constant.KeyOauth2Enabled) ||
		(r.bool(constant.KeyOidcEnabled) && r.get(constant.KeyOidcType) == constant.SSOTypeOAuth2)

	if oauth2Enabled {
		clientID := r.get(constant.KeyOauth2ClientID, constant.KeyOidcClientID)
		authURL := r.get(constant.KeyOauth2AuthURL)
		tokenURL := r.get(constant.KeyOauth2TokenURL)
		userinfoURL := r.get(constant.KeyOauth2UserinfoURL)
		if clientID != "" && authURL != "" && tokenURL != "" && userinfoURL != "" {
			providers = append(providers, gin.H{
				"id":        "oauth2",
				"type":      constant.SSOTypeOAuth2,
				"name":      r.getWithDefault("OAuth 2.0", constant.KeyOauth2Name),
				"login_url": "/api/v1/auth/oauth2/login",
			})
		}
	}

	if len(providers) == 0 {
		return gin.H{"enabled": false, "providers": providers}
	}

	// 兼容单通道旧字段回显
	return gin.H{
		"enabled":   true,
		"providers": providers,
		"type":      providers[0]["type"],
		"name":      providers[0]["name"],
	}
}

// getOidcConfig 读取标准 OIDC 配置
func (ac *AuthController) getOidcConfig(c *gin.Context) *oidcConfig {
	r := securityReader{ac.settingsService}
	if !r.bool(constant.KeyOidcEnabled) {
		return nil
	}

	clientID := r.get(constant.KeyOidcClientID)
	issuer := NormalizeIssuer(r.get(constant.KeyOidcIssuer))
	if clientID == "" || issuer == "" {
		return nil
	}

	return &oidcConfig{
		Type:         constant.SSOTypeOIDC,
		Issuer:       issuer,
		ClientID:     clientID,
		ClientSecret: r.secret(constant.KeyOidcClientSecret),
		RedirectURL:  ac.resolveRedirectURL(r.get(constant.KeyOidcRedirectURL), "/api/v1/auth/oidc/callback", c),
		Scopes:       r.scopes(r.get(constant.KeyOidcScopes), oidc.ScopeOpenID, "profile", "email"),
		AutoCreate:   r.bool(constant.KeyOidcAutoCreate),
		DefaultRole:  r.role(constant.KeyOidcDefaultRole),
	}
}

// getOAuth2Config 读取通用 OAuth 2.0 配置
func (ac *AuthController) getOAuth2Config(c *gin.Context) *oidcConfig {
	r := securityReader{ac.settingsService}
	enabled := r.bool(constant.KeyOauth2Enabled) ||
		(r.bool(constant.KeyOidcEnabled) && r.get(constant.KeyOidcType) == constant.SSOTypeOAuth2)
	if !enabled {
		return nil
	}

	clientID := r.get(constant.KeyOauth2ClientID, constant.KeyOidcClientID)
	authURL := r.get(constant.KeyOauth2AuthURL)
	tokenURL := r.get(constant.KeyOauth2TokenURL)
	userinfoURL := r.get(constant.KeyOauth2UserinfoURL)
	if clientID == "" || authURL == "" || tokenURL == "" || userinfoURL == "" {
		return nil
	}

	return &oidcConfig{
		Type:         constant.SSOTypeOAuth2,
		ClientID:     clientID,
		ClientSecret: r.secret(constant.KeyOauth2ClientSecret, constant.KeyOidcClientSecret),
		RedirectURL:  ac.resolveRedirectURL("", "/api/v1/auth/oauth2/callback", c),
		Scopes:       r.scopes(r.get(constant.KeyOauth2Scopes, constant.KeyOidcScopes)),
		AutoCreate:   r.bool(constant.KeyOauth2AutoCreate, constant.KeyOidcAutoCreate),
		DefaultRole:  r.role(constant.KeyOauth2DefaultRole, constant.KeyOidcDefaultRole),
		AuthURL:      authURL,
		TokenURL:     tokenURL,
		UserinfoURL:  userinfoURL,
		DisplayName:  r.getWithDefault("OAuth 2.0", constant.KeyOauth2Name),
		UserKey:      r.get(constant.KeyOauth2UserKey),
	}
}

// oidcConfig 读取并校验单点登录配置（智能适配通道，兼容旧调用）
func (ac *AuthController) oidcConfig(ctx ...*gin.Context) *oidcConfig {
	var c *gin.Context
	if len(ctx) > 0 {
		c = ctx[0]
	}

	if c != nil && c.Request != nil {
		path := c.Request.URL.Path
		providerParam := c.Query("provider")
		if strings.Contains(path, "/oauth2") || providerParam == constant.SSOTypeOAuth2 {
			if cfg := ac.getOAuth2Config(c); cfg != nil {
				return cfg
			}
		}
		if strings.Contains(path, "/oidc") || providerParam == constant.SSOTypeOIDC {
			if cfg := ac.getOidcConfig(c); cfg != nil {
				return cfg
			}
		}
	}

	// 默认探测：若 OIDC 开启则走 OIDC；否则若 OAuth2 开启则走 OAuth2
	if cfg := ac.getOidcConfig(c); cfg != nil {
		return cfg
	}
	return ac.getOAuth2Config(c)
}

// newSSOHTTPClient 创建统一的 SSO HTTP 客户端（支持私有自签名证书与内网代理环境）
func newSSOHTTPClient() *http.Client {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 兼容自建与私有部署环境的自签名证书与内网 IP
		},
	}
	return &http.Client{
		Transport: tr,
		Timeout:   15 * time.Second,
	}
}

// oidcProviderFor 返回可用的 provider / verifier / oauth2 配置（动态组装避免脏缓存）
func (ac *AuthController) oidcProviderFor(ctx context.Context, cfg *oidcConfig) (*oidc.IDTokenVerifier, *oauth2.Config, error) {
	// 通用 OAuth 2.0 模式：直接构建标准 oauth2.Config，不需要 OIDC Provider 与 JWKS 验签器
	if cfg.Type == constant.SSOTypeOAuth2 {
		oauthCfg := &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.AuthURL,
				TokenURL: cfg.TokenURL,
			},
			RedirectURL: cfg.RedirectURL,
			Scopes:      cfg.Scopes,
		}
		return nil, oauthCfg, nil
	}

	var provider *oidc.Provider
	var verifier *oidc.IDTokenVerifier

	oidcCache.mu.Lock()
	if oidcCache.provider != nil && oidcCache.issuer == cfg.Issuer && oidcCache.clientID == cfg.ClientID {
		provider = oidcCache.provider
		verifier = oidcCache.verifier
		oidcCache.mu.Unlock()
	} else {
		oidcCache.mu.Unlock()

		reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		httpClient := newSSOHTTPClient()
		reqCtx = oidc.ClientContext(reqCtx, httpClient)
		reqCtx = context.WithValue(reqCtx, oauth2.HTTPClient, httpClient)

		// 优先使用 InsecureIssuerURLContext 允许用户输入的 Issuer 与发现文档宽松匹配
		issuerCtx := oidc.InsecureIssuerURLContext(reqCtx, cfg.Issuer)

		p, err := oidc.NewProvider(issuerCtx, cfg.Issuer)
		if err != nil {
			// 若因末尾斜杠不一致（例如发现文档定义了尾部斜杠），自动切换斜杠重试
			altIssuer := strings.TrimRight(cfg.Issuer, "/")
			if altIssuer == cfg.Issuer {
				altIssuer = cfg.Issuer + "/"
			}
			altCtx := oidc.InsecureIssuerURLContext(reqCtx, altIssuer)
			if p2, err2 := oidc.NewProvider(altCtx, altIssuer); err2 == nil {
				p = p2
				err = nil
			}
		}
		if err != nil {
			return nil, nil, err
		}

		provider = p
		verifier = p.Verifier(&oidc.Config{ClientID: cfg.ClientID})

		oidcCache.mu.Lock()
		oidcCache.provider = provider
		oidcCache.verifier = verifier
		oidcCache.issuer = cfg.Issuer
		oidcCache.clientID = cfg.ClientID
		oidcCache.mu.Unlock()
	}

	// 始终根据当前请求动态构造 oauth2.Config（保证多域名/反代/更换密钥时 100% 精确生效）
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
	}

	return verifier, oauthCfg, nil
}

type oidcClaims struct {
	Sub               string   `json:"sub"`
	Email             string   `json:"email"`
	PreferredUsername string   `json:"preferred_username"`
	Name              string   `json:"name"`
	Groups            []string `json:"groups"`
}

// 确保 services 包被引用（settingsService 类型来自该包）
var _ = services.NewSettingsService
