package cache

import (
	"testing"
	"time"
)

func TestPublicSettingsCache(t *testing.T) {
	// 确保测试初始状态干净
	InvalidatePublicSettingsCache()

	// 1. 初始未命中测试
	if val, ok := GetPublicSettingsCache(); ok || val != nil {
		t.Fatalf("期望缓存未命中，实际返回: %v, ok=%v", val, ok)
	}

	// 2. 写入缓存并读取验证 (0ms 命中)
	mockData := map[string]any{
		"title":     "测试面板",
		"subtitle":  "测试副标题",
		"demo_mode": false,
	}
	SetPublicSettingsCache(mockData)

	val, ok := GetPublicSettingsCache()
	if !ok || val == nil {
		t.Fatalf("期望缓存命中，实际未命中")
	}
	resMap, isMap := val.(map[string]any)
	if !isMap || resMap["title"] != "测试面板" {
		t.Fatalf("缓存内容不匹配: %+v", val)
	}

	// 3. 主动失效测试 (0ms 实时刷新)
	InvalidatePublicSettingsCache()
	if val, ok := GetPublicSettingsCache(); ok || val != nil {
		t.Fatalf("InvalidatePublicSettingsCache 后期望立即未命中，实际: %v", val)
	}

	// 4. 模拟 30s 兜底自动过期
	SetPublicSettingsCache(mockData)
	publicSettingsMu.Lock()
	// 人工将缓存时间调整为 31 秒之前，模拟超时
	publicSettingsCacheAt = time.Now().Add(-31 * time.Second)
	publicSettingsMu.Unlock()

	if val, ok := GetPublicSettingsCache(); ok || val != nil {
		t.Fatalf("超过 30s 兜底时间后期望缓存自动失效，实际仍然返回: %v", val)
	}
}
