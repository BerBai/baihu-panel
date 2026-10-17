package controllers

import (
	"sync"

	"github.com/engigu/baihu-panel/internal/cache"
	"github.com/engigu/baihu-panel/internal/constant"
	"github.com/engigu/baihu-panel/internal/oidc"
	"github.com/engigu/baihu-panel/internal/services"
	"github.com/engigu/baihu-panel/internal/utils"
	"github.com/gin-gonic/gin"
)

// PublicController 负责处理系统所有免密/公开只读信息接口
type PublicController struct {
	settingsService *services.SettingsService
	fetchMu         sync.Mutex
}

// NewPublicController 实例化公共接口控制器
func NewPublicController(settingsService *services.SettingsService) *PublicController {
	return &PublicController{
		settingsService: settingsService,
	}
}

// InvalidateCache 主动使公共配置缓存失效
func (pc *PublicController) InvalidateCache() {
	cache.InvalidatePublicSettingsCache()
}

// GetPublicSettings 获取公开的站点元数据与 SSO 登录状态（无需认证，登录页与基础 UI 专用）
func (pc *PublicController) GetPublicSettings(c *gin.Context) {
	// 1. 优先命中 30s 内存缓存（耗时 <0.05ms，彻底阻断同一时刻/高频并发对 DB 的穿透）
	if cached, ok := cache.GetPublicSettingsCache(); ok && cached != nil {
		utils.Success(c, cached)
		return
	}

	// 2. 加互斥锁进行双检加载，防止缓存过期瞬时的高并发重复穿透重算
	pc.fetchMu.Lock()
	defer pc.fetchMu.Unlock()

	if cached, ok := cache.GetPublicSettingsCache(); ok && cached != nil {
		utils.Success(c, cached)
		return
	}

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

	res := gin.H{
		constant.KeyTitle:    title,
		constant.KeySubtitle: subtitle,
		constant.KeyIcon:     icon,
		"demo_mode":          constant.DemoMode,
		"sso":                ssoStatus,
	}

	// 写入常驻内存缓存（30s 兜底自动过期；面板修改时 0ms 瞬间主动失效）
	cache.SetPublicSettingsCache(res)

	utils.Success(c, res)
}
