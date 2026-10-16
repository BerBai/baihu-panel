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
	oauth    *oauth2.Config
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
func (ac *AuthController) resolveRedirectURL(custom string, c *gin.Context) string {
	if trimmed := strings.TrimSpace(custom); trimmed != "" {
		return trimmed
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
			return fmt.Sprintf("%s://%s/api/v1/auth/oidc/callback", scheme, host)
		}
	}
	return "/api/v1/auth/oidc/callback"
}

// oidcConfig 读取并校验单点登录（OIDC / OAuth 2.0）设置；未启用或配置不全时返回 nil
func (ac *AuthController) oidcConfig(ctx ...*gin.Context) *oidcConfig {
	if ac.settingsService.Get(constant.SectionSecurity, constant.KeyOidcEnabled) != "true" {
		return nil
	}

	ssoType := ac.settingsService.Get(constant.SectionSecurity, constant.KeyOidcType)
	if ssoType != constant.SSOTypeOAuth2 {
		ssoType = constant.SSOTypeOIDC
	}

	clientID := strings.TrimSpace(ac.settingsService.Get(constant.SectionSecurity, constant.KeyOidcClientID))
	if clientID == "" {
		return nil
	}
	clientSecret := utils.DecryptSecret(ac.settingsService.Get(constant.SectionSecurity, constant.KeyOidcClientSecret))

	var c *gin.Context
	if len(ctx) > 0 {
		c = ctx[0]
	}
	redirect := ac.resolveRedirectURL(ac.settingsService.Get(constant.SectionSecurity, constant.KeyOidcRedirectURL), c)

	role := ac.settingsService.Get(constant.SectionSecurity, constant.KeyOidcDefaultRole)
	if role != constant.AdminRole {
		role = constant.DefaultRole
	}
	autoCreate := ac.settingsService.Get(constant.SectionSecurity, constant.KeyOidcAutoCreate) == "true"

	if ssoType == constant.SSOTypeOIDC {
		issuer := strings.TrimRight(ac.settingsService.Get(constant.SectionSecurity, constant.KeyOidcIssuer), "/")
		if issuer == "" {
			return nil
		}

		// 允许管理员追加 scope（逗号或空格分隔），始终包含 openid
		scopes := []string{oidc.ScopeOpenID, "profile", "email"}
		for _, s := range strings.FieldsFunc(ac.settingsService.Get(constant.SectionSecurity, constant.KeyOidcScopes), func(r rune) bool {
			return r == ',' || r == ' ' || r == '\n'
		}) {
			if s = strings.TrimSpace(s); s != "" && s != oidc.ScopeOpenID {
				scopes = append(scopes, s)
			}
		}

		return &oidcConfig{
			Type:         constant.SSOTypeOIDC,
			Issuer:       issuer,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirect,
			Scopes:       scopes,
			AutoCreate:   autoCreate,
			DefaultRole:  role,
		}
	}

	// 通用 OAuth 2.0 模式
	authURL := strings.TrimSpace(ac.settingsService.Get(constant.SectionSecurity, constant.KeyOauth2AuthURL))
	tokenURL := strings.TrimSpace(ac.settingsService.Get(constant.SectionSecurity, constant.KeyOauth2TokenURL))
	userinfoURL := strings.TrimSpace(ac.settingsService.Get(constant.SectionSecurity, constant.KeyOauth2UserinfoURL))
	if authURL == "" || tokenURL == "" || userinfoURL == "" {
		return nil
	}

	var scopes []string
	for _, s := range strings.FieldsFunc(ac.settingsService.Get(constant.SectionSecurity, constant.KeyOidcScopes), func(r rune) bool {
		return r == ',' || r == ' ' || r == '\n'
	}) {
		if s = strings.TrimSpace(s); s != "" {
			scopes = append(scopes, s)
		}
	}

	displayName := strings.TrimSpace(ac.settingsService.Get(constant.SectionSecurity, constant.KeyOauth2Name))
	userKey := strings.TrimSpace(ac.settingsService.Get(constant.SectionSecurity, constant.KeyOauth2UserKey))

	return &oidcConfig{
		Type:         constant.SSOTypeOAuth2,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirect,
		Scopes:       scopes,
		AutoCreate:   autoCreate,
		DefaultRole:  role,
		AuthURL:      authURL,
		TokenURL:     tokenURL,
		UserinfoURL:  userinfoURL,
		DisplayName:  displayName,
		UserKey:      userKey,
	}
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

// oidcProviderFor 返回可用的 provider / verifier / oauth2 配置（设置变更后自动重建）
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

	oidcCache.mu.Lock()
	defer oidcCache.mu.Unlock()

	if oidcCache.provider != nil && oidcCache.issuer == cfg.Issuer && oidcCache.clientID == cfg.ClientID {
		return oidcCache.verifier, oidcCache.oauth, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	httpClient := newSSOHTTPClient()
	ctx = oidc.ClientContext(ctx, httpClient)
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	// 优先使用 InsecureIssuerURLContext 允许用户输入的 Issuer 与发现文档宽松匹配
	issuerCtx := oidc.InsecureIssuerURLContext(ctx, cfg.Issuer)

	provider, err := oidc.NewProvider(issuerCtx, cfg.Issuer)
	if err != nil {
		// 若因末尾斜杠不一致（例如发现文档定义了尾部斜杠），自动切换斜杠重试
		altIssuer := strings.TrimRight(cfg.Issuer, "/")
		if altIssuer == cfg.Issuer {
			altIssuer = cfg.Issuer + "/"
		}
		altCtx := oidc.InsecureIssuerURLContext(ctx, altIssuer)
		if p2, err2 := oidc.NewProvider(altCtx, altIssuer); err2 == nil {
			provider = p2
			err = nil
		}
	}
	if err != nil {
		return nil, nil, err
	}

	oidcCache.provider = provider
	oidcCache.verifier = provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
	oidcCache.oauth = &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
	}
	oidcCache.issuer = cfg.Issuer
	oidcCache.clientID = cfg.ClientID

	return oidcCache.verifier, oidcCache.oauth, nil
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
