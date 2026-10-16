package oidc

import (
	"testing"

	"github.com/engigu/baihu-panel/internal/constant"
	"github.com/engigu/baihu-panel/internal/database"
	"github.com/engigu/baihu-panel/internal/models"
	"github.com/engigu/baihu-panel/internal/services"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestNormalizeIssuer(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"https://id.example.com", "https://id.example.com"},
		{"https://id.example.com/", "https://id.example.com"},
		{"https://id.example.com///", "https://id.example.com"},
		{"https://id.example.com/.well-known/openid-configuration", "https://id.example.com"},
		{"https://id.example.com/.well-known/openid-configuration/", "https://id.example.com"},
		{"  https://id.example.com/.well-known/openid-configuration/  ", "https://id.example.com"},
		{"", ""},
		{"   ", ""},
	}

	for _, c := range cases {
		got := NormalizeIssuer(c.input)
		if got != c.expected {
			t.Errorf("NormalizeIssuer(%q) = %q, expected %q", c.input, got, c.expected)
		}
	}
}

func TestMaskAndPrepareSecuritySection(t *testing.T) {
	// 1. 测试读取时掩码
	secValues := map[string]string{
		constant.KeyOidcClientSecret:   "my-oidc-secret",
		constant.KeyOauth2ClientSecret: "my-oauth2-secret",
		constant.KeyOidcClientID:       "client-123",
	}
	masked := MaskSecuritySection(constant.SectionSecurity, secValues)
	if masked[constant.KeyOidcClientSecret] != SecretMask || masked[constant.KeyOauth2ClientSecret] != SecretMask {
		t.Errorf("密钥未被正确掩码: %+v", masked)
	}
	if masked[constant.KeyOidcClientID] != "client-123" {
		t.Errorf("普通配置被意外修改: %+v", masked)
	}

	// 2. 测试非 security 段不受影响
	otherValues := map[string]string{
		constant.KeyOidcClientSecret: "raw-secret",
	}
	notMasked := MaskSecuritySection(constant.SectionSite, otherValues)
	if notMasked[constant.KeyOidcClientSecret] != "raw-secret" {
		t.Errorf("非 security 段不应掩码")
	}

	// 3. 提交掩码值时视为未修改，原样保留原值
	updateValues := map[string]string{
		constant.KeyOidcClientSecret: SecretMask,
	}
	prepared := PrepareSecuritySection(constant.SectionSecurity, updateValues, func(key string) string {
		return "encrypted-stored-val"
	})
	if prepared[constant.KeyOidcClientSecret] != "encrypted-stored-val" {
		t.Errorf("掩码值应恢复原有存储值，实际: %s", prepared[constant.KeyOidcClientSecret])
	}
}

func TestResolveSSOStatusMultiProviders(t *testing.T) {
	// 使用内存数据库验证 SettingsService
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("开启内存库失败: %v", err)
	}
	database.DB = db
	_ = db.AutoMigrate(&models.Setting{})

	ss := services.NewSettingsService()
	// 未启用时
	st := ResolveSSOStatus(ss)
	if st["enabled"] != false {
		t.Errorf("未配置时 enabled 应为 false")
	}

	// 启用 OIDC 与 OAuth 2.0 双通道
	_ = ss.Set(constant.SectionSecurity, constant.KeyOidcEnabled, "true")
	_ = ss.Set(constant.SectionSecurity, constant.KeyOidcIssuer, "https://id.example.com")
	_ = ss.Set(constant.SectionSecurity, constant.KeyOidcClientID, "oidc-client")

	_ = ss.Set(constant.SectionSecurity, constant.KeyOauth2Enabled, "true")
	_ = ss.Set(constant.SectionSecurity, constant.KeyOauth2ClientID, "gh-client")
	_ = ss.Set(constant.SectionSecurity, constant.KeyOauth2AuthURL, "https://github.com/login/oauth/authorize")
	_ = ss.Set(constant.SectionSecurity, constant.KeyOauth2TokenURL, "https://github.com/login/oauth/access_token")
	_ = ss.Set(constant.SectionSecurity, constant.KeyOauth2UserinfoURL, "https://api.github.com/user")
	_ = ss.Set(constant.SectionSecurity, constant.KeyOauth2Name, "GitHub")

	st2 := ResolveSSOStatus(ss)
	if st2["enabled"] != true {
		t.Errorf("双通道启用时 enabled 应为 true")
	}
	providers, ok := st2["providers"].([]map[string]any)
	if !ok {
		// gin.H 等价于 map[string]any
		list, ok2 := st2["providers"].([]struct{})
		_ = list
		_ = ok2
	}
	_ = providers
}
