package controllers

import (
	"github.com/engigu/baihu-panel/internal/constant"
	"github.com/engigu/baihu-panel/internal/oidc"
	"github.com/engigu/baihu-panel/internal/services"
	"github.com/engigu/baihu-panel/internal/utils"
	"github.com/gin-gonic/gin"
)

// PublicController 负责处理系统所有免密/公开只读信息接口
type PublicController struct {
	settingsService *services.SettingsService
}

// NewPublicController 实例化公共接口控制器
func NewPublicController(settingsService *services.SettingsService) *PublicController {
	return &PublicController{
		settingsService: settingsService,
	}
}

// GetPublicSettings 获取公开的站点元数据与 SSO 登录状态（无需认证，登录页与基础 UI 专用）
func (pc *PublicController) GetPublicSettings(c *gin.Context) {
	settings := pc.settingsService.GetSection(constant.SectionSite)

	title := settings[constant.KeyTitle]
	if title == "" {
		title = "白虎面板"
	}
	subtitle := settings[constant.KeySubtitle]
	if subtitle == "" {
		subtitle = "极致轻量、高性能的自动化任务调度平台"
	}
	icon := settings[constant.KeyIcon]
	if icon == "" {
		icon = constant.DefaultIcon
	}

	// 提取轻量无敏感信息的 SSO 状态（type、name、enabled）
	ssoStatus := oidc.ResolveSSOStatus(pc.settingsService)

	utils.Success(c, gin.H{
		constant.KeyTitle:    title,
		constant.KeySubtitle: subtitle,
		constant.KeyIcon:     icon,
		"demo_mode":          constant.DemoMode,
		"sso":                ssoStatus,
	})
}
