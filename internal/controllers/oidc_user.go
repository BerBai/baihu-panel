package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/engigu/baihu-panel/internal/constant"
	"github.com/engigu/baihu-panel/internal/database"
	"github.com/engigu/baihu-panel/internal/eventbus"
	"github.com/engigu/baihu-panel/internal/models"
	"github.com/engigu/baihu-panel/internal/utils"

	"github.com/gin-gonic/gin"
)

func (ac *AuthController) logOidcFailure(c *gin.Context, username, message string) {
	eventbus.DefaultBus.Publish(eventbus.Event{
		Type: constant.EventUserLogin,
		Payload: map[string]interface{}{
			"ip":        c.ClientIP(),
			"username":  username,
			"userAgent": c.GetHeader("User-Agent"),
			"status":    "failed",
			"message":   "单点登录失败：" + message,
		},
	})
}

// resolveOidcUser 按 subject 精确匹配 → 邮箱绑定 → 按配置自动创建
func (ac *AuthController) resolveOidcUser(cfg *oidcConfig, claims *oidcClaims) (*models.User, bool) {
	if claims.Sub == "" {
		return nil, false
	}

	var user models.User
	res := database.DB.Where("oidc_subject = ?", claims.Sub).Limit(1).Find(&user)
	if res.Error == nil && res.RowsAffected > 0 {
		return &user, false
	}

	// 兼容已存在账号：邮箱相同且尚未绑定过任何 subject，则自动绑定
	if claims.Email != "" {
		res := database.DB.Where("email = ? AND (oidc_subject = '' OR oidc_subject IS NULL)", claims.Email).Limit(1).Find(&user)
		if res.Error == nil && res.RowsAffected > 0 {
			database.DB.Model(&user).Update("oidc_subject", claims.Sub)
			return &user, false
		}
	}

	if !cfg.AutoCreate {
		return nil, false
	}

	username := claims.PreferredUsername
	if username == "" {
		username = claims.Email
	}
	if username == "" {
		username = "sso_" + claims.Sub
	}

	created := ac.userService.CreateOidcUser(ac.uniqueUsername(username), claims.Email, cfg.DefaultRole, claims.Sub)
	if created == nil {
		return nil, false
	}
	return created, true
}

func (ac *AuthController) uniqueUsername(base string) string {
	candidate := base
	for i := 0; i < 20; i++ {
		var count int64
		database.DB.Model(&models.User{}).Where("username = ?", candidate).Count(&count)
		if count == 0 {
			return candidate
		}
		candidate = base + "_" + utils.RandomString(4)
	}
	return base + "_" + utils.RandomString(8)
}

// fetchOAuth2Userinfo 通过 Access Token 调用第三方 Userinfo 端点获取用户信息
func (ac *AuthController) fetchOAuth2Userinfo(ctx context.Context, cfg *oidcConfig, accessToken string) (*oidcClaims, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserinfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建用户信息请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Baihu-Panel-SSO/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求用户信息接口失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("用户信息接口响应状态异常 (%d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, fmt.Errorf("读取用户信息失败: %w", err)
	}

	claims, err := parseOAuth2Userinfo(body, cfg.UserKey)
	if err != nil {
		return nil, err
	}

	// 针对 GitHub 等平台：若未公开邮箱，尝试请求 /user/emails 补齐主邮箱
	if claims.Email == "" && strings.Contains(cfg.UserinfoURL, "api.github.com") {
		if ghEmail := fetchGitHubPrimaryEmail(ctx, accessToken); ghEmail != "" {
			claims.Email = ghEmail
		}
	}

	return claims, nil
}

// parseOAuth2Userinfo 从任意 OAuth 2.0 用户信息 JSON 响应中智能提取身份字段
func parseOAuth2Userinfo(raw []byte, customUserKey string) (*oidcClaims, error) {
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("解析用户信息 JSON 失败: %w", err)
	}

	claims := &oidcClaims{}

	// 1. 提取唯一标识 Sub
	if customUserKey != "" {
		if val, ok := data[customUserKey]; ok && val != nil {
			claims.Sub = strings.TrimSpace(fmt.Sprintf("%v", val))
		}
	}
	if claims.Sub == "" {
		candidates := []string{"sub", "id", "user_id", "uid", "openid", "login", "username"}
		for _, key := range candidates {
			if val, ok := data[key]; ok && val != nil {
				str := strings.TrimSpace(fmt.Sprintf("%v", val))
				if str != "" && str != "<nil>" {
					claims.Sub = str
					break
				}
			}
		}
	}
	if claims.Sub == "" {
		return nil, fmt.Errorf("未能从用户信息中识别出唯一身份标识 (ID)")
	}

	// 2. 提取用户名
	userCandidates := []string{"login", "username", "preferred_username", "name", "nickname"}
	for _, key := range userCandidates {
		if val, ok := data[key]; ok && val != nil {
			str := strings.TrimSpace(fmt.Sprintf("%v", val))
			if str != "" && str != "<nil>" {
				claims.PreferredUsername = str
				break
			}
		}
	}
	if claims.PreferredUsername == "" {
		claims.PreferredUsername = claims.Sub
	}

	// 3. 提取邮箱
	emailCandidates := []string{"email", "mail", "user_email"}
	for _, key := range emailCandidates {
		if val, ok := data[key]; ok && val != nil {
			str := strings.TrimSpace(fmt.Sprintf("%v", val))
			if str != "" && strings.Contains(str, "@") {
				claims.Email = str
				break
			}
		}
	}

	// 4. 提取显示名称
	if val, ok := data["name"]; ok && val != nil {
		claims.Name = fmt.Sprintf("%v", val)
	}

	return claims, nil
}

func fetchGitHubPrimaryEmail(ctx context.Context, token string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Baihu-Panel-SSO/1.0")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return ""
	}
	for _, e := range emails {
		if e.Primary && e.Verified && e.Email != "" {
			return e.Email
		}
	}
	if len(emails) > 0 {
		return emails[0].Email
	}
	return ""
}
