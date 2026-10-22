package config

import (
	"os"
	"path/filepath"
	"testing"
)

// newModelFile 在临时目录创建假模型文件，返回其路径（Validate 的 fs 检查用）。
func newModelFile(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "model.ninfer")
	if err := os.WriteFile(p, []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// Test_DefaultConfig_when_batRecipeApplied_returnsAll33Defaults
// 断言 DefaultConfig() 的 32 个字段逐一等于 start-ninfer.bat 默认配方
// （整结构相等比较，无字段遗漏）。
func Test_DefaultConfig_when_batRecipeApplied_returnsAll33Defaults(t *testing.T) {
	got := DefaultConfig()

	want := Config{
		// 服务（3）
		ListenHost:  "127.0.0.1",
		Port:        8080,
		Concurrency: 1,
		// 模型（3）
		ModelPath:   "Qwen3.8-27B-GSQ-RCO-IQ3_S-ninfer-v3.ninfer",
		Gpu:         "0",
		StageLayers: "",
		// 上下文与内存（6）
		Context:   131072,
		Budget:    36864,
		Reserve:   16384,
		HostMiB:   12288,
		MaxTokens: 16384,
		Prefill:   512,
		// KV 缓存（5）
		KvType:           "int8",
		SinkTokens:       12288,
		RecentTokens:     1024,
		RetainedSessions: 1,
		FastPrefill:      false,
		// 推测解码（4）
		MtpDrafts:     4,
		NgramDrafts:   31,
		NgramMinMatch: 12,
		AdaptiveMtp:   true,
		// 视觉（3）
		Vision:          true,
		VisionResidency: "cpu",
		VisionTokens:    1024,
		// 思考与设备（2）
		ThinkingBudget: 8192,
		DeviceProfile:  "off",
		// SSD 分层（2）
		DiskPath: `D:\kv-cache`,
		DiskMiB:  4096,
		// 检索调优（4）
		NoLrfu:           false,
		LrfuLambda:       0.5,
		LrfuTopFraction:  0.9,
		ScoreAggregation: "max",
	}

	if got != want {
		t.Fatalf("DefaultConfig() = %+v, want %+v", got, want)
	}
}
