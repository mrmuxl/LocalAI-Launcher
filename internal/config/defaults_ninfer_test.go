package config

import "testing"

// Test_DefaultConfig_when_fieldTableApplied_returnsAll34Defaults
// 断言 DefaultConfig() 的 34 个字段逐一等于计划「字段表」默认值
// （整结构相等比较，无字段遗漏；空串/false/0 字段按零值断言）。
func Test_DefaultConfig_when_fieldTableApplied_returnsAll34Defaults(t *testing.T) {
	got := DefaultConfig()

	want := Config{
		// 服务（4）
		Host:     "127.0.0.1",
		Port:     8080,
		LogLevel: "info",
		Workers:  4,
		// 模型（5）
		ModelName: "qwen3-27b",
		ModelPath: "",
		CtxSize:   32768,
		Threads:   0,
		Quant:     "q4_k_m",
		// 采样（5）
		Temperature:   0.7,
		TopP:          0.9,
		TopK:          40,
		RepeatPenalty: 1.1,
		MaxTokens:     4096,
		// 网络（4）
		APIKey:      "",
		CORSOrigins: "*",
		TimeoutSec:  300,
		MaxConns:    32,
		// 安全（3）
		TLS:      false,
		CertPath: "",
		KeyPath:  "",
		// 观测（4）
		MetricsEnabled: true,
		MetricsPath:    "/metrics",
		LogFile:        "",
		AccessLog:      true,
		// 资源（5）
		MemLimitGB:   0,
		SwapGB:       0,
		NumaNode:     -1,
		Priority:     "normal",
		AffinityMask: "",
		// 缓存（4）
		CacheDir:    "./cache",
		CacheSizeMB: 1024,
		EvictPolicy: "lru",
		Prefetch:    false,
	}

	if got != want {
		t.Errorf("DefaultConfig() = %+v\nwant:        %+v", got, want)
	}
}
