package cache

import (
	"sync"
	"time"
)

// PublicSettingsTTL 公共设置内存缓存兜底过期时间（30秒）
// 兼顾高并发防击穿与运维手动直改数据库时的自动感知与刷新
const PublicSettingsTTL = 30 * time.Second

var (
	publicSettingsCache   any
	publicSettingsCacheAt time.Time
	publicSettingsMu      sync.RWMutex
)

// GetPublicSettingsCache 获取有效的公共设置缓存
// 若缓存未命中或已超过 30 秒兜底 TTL，返回 (nil, false)
func GetPublicSettingsCache() (any, bool) {
	publicSettingsMu.RLock()
	defer publicSettingsMu.RUnlock()

	if publicSettingsCache != nil && time.Since(publicSettingsCacheAt) < PublicSettingsTTL {
		return publicSettingsCache, true
	}
	return nil, false
}

// SetPublicSettingsCache 写入/刷新公共设置内存缓存并记录时间戳
func SetPublicSettingsCache(data any) {
	publicSettingsMu.Lock()
	defer publicSettingsMu.Unlock()

	publicSettingsCache = data
	publicSettingsCacheAt = time.Now()
}

// InvalidatePublicSettingsCache 主动使公共设置缓存失效
// 当面板在后台修改了站点配置或安全/SSO设置时立即调用，实现 0ms 瞬间刷新
func InvalidatePublicSettingsCache() {
	publicSettingsMu.Lock()
	defer publicSettingsMu.Unlock()

	publicSettingsCache = nil
	publicSettingsCacheAt = time.Time{}
}
