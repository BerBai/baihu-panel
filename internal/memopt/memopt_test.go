package memopt

import (
	"testing"

	"github.com/engigu/baihu-panel/internal/constant"
)

func TestMemOpt_FreeAndMetrics(t *testing.T) {
	Init()

	initialRSS := GetRSS()

	// 模拟一次大内存占用
	buf := make([]byte, 30*1024*1024)
	for i := 0; i < len(buf); i += 4096 {
		buf[i] = 1
	}

	inflatedRSS := GetRSS()
	t.Logf("内存分配后 RSS: %d, 较初始增长约 %.2f MB", inflatedRSS, float64(inflatedRSS-initialRSS)/1024/1024)

	// 释放引用并执行 Checkpoint
	buf = nil
	_ = buf
	Checkpoint("test_phase")

	afterRSS := GetRSS()
	t.Logf("Checkpoint 释放后 RSS: %d, 较峰值下降约 %.2f MB", afterRSS, float64(inflatedRSS-afterRSS)/1024/1024)

	metrics := GetMetrics()
	if metrics.NumGoroutine <= 0 {
		t.Errorf("NumGoroutine 异常: %d", metrics.NumGoroutine)
	}

	formatted := FormatBytes(1024 * 1024 * 5)
	if formatted != "5.0 MB" {
		t.Errorf("FormatBytes 结果不符预期: %s", formatted)
	}
}

func TestGetContainerTrimTargets_ExcludesBaseRuntime(t *testing.T) {
	targets := GetContainerTrimTargets()
	for _, target := range targets {
		if target == "/usr/local/bin/mise" || target == "/app/baihu" || target == constant.ContainerMiseBaseDir {
			t.Errorf("日常清理目标中不应包含基础运行环境底座: %s", target)
		}
	}

	startupTargets := GetContainerStartupTrimTargets()
	// 无论非 Docker 还是 Docker，Startup 目标数量都应 >= 日常目标
	if len(startupTargets) < len(targets) {
		t.Errorf("启动就绪全量目标数量 (%d) 应 >= 日常目标 (%d)", len(startupTargets), len(targets))
	}
}

func TestOnTaskFinished_SafetyAndCooldown(t *testing.T) {
	// 非 Docker 环境下安全退出，不发生 panic
	OnTaskFinished("")
	OnTaskFinished("/")
	OnTaskFinished("/etc")
	OnTaskFinished(constant.ScriptsWorkDir + "/my_task")
}

