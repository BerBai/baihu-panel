package controllers

import (
	"github.com/engigu/baihu-panel/internal/constant"
	"github.com/engigu/baihu-panel/internal/utils"
)

// oidcSecretMask 前端回显用的掩码；提交同名值时视为「未修改」，保留库中原值
const oidcSecretMask = "********"

// maskOidcSection 读取 security 段时掩掉客户端密钥，避免密钥回显到浏览器
func maskOidcSection(section string, values map[string]string) map[string]string {
	if section != constant.SectionSecurity || values == nil {
		return values
	}
	if v, ok := values[constant.KeyOidcClientSecret]; ok && v != "" {
		values[constant.KeyOidcClientSecret] = oidcSecretMask
	}
	if v, ok := values[constant.KeyOauth2ClientSecret]; ok && v != "" {
		values[constant.KeyOauth2ClientSecret] = oidcSecretMask
	}
	return values
}

// prepareOidcSection 写入 security 段前处理客户端密钥：
// 掩码值 = 保留原值；明文 = 加密后落库；空串 = 清空
func prepareOidcSection(section string, values map[string]string, getter func(key string) string) map[string]string {
	if section != constant.SectionSecurity || values == nil {
		return values
	}
	for _, key := range []string{constant.KeyOidcClientSecret, constant.KeyOauth2ClientSecret} {
		raw, ok := values[key]
		if !ok {
			continue
		}
		switch raw {
		case oidcSecretMask:
			// 未修改：原样写回当前值（可能是密文），避免把掩码存进库
			if getter != nil {
				values[key] = getter(key)
			}
		case "":
			values[key] = ""
		default:
			if encrypted, err := utils.Encrypt(raw); err == nil {
				values[key] = encrypted
			}
			// 未配置主密钥时 Encrypt 报错，此时按明文存储，DecryptSecret 会在读取时原样返回
		}
	}
	return values
}
