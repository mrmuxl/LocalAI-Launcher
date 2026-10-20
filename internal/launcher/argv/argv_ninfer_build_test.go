package argv

import (
	"strings"
	"testing"

	"ninferrmanager/internal/launcher/config"
)

// TestBuildArgs 用表驱动用例锁定计划「argv 构造契约」：
// 默认配置 + 4 个变体，输出 slice 逐元素精确相等（不使用 Contains 类弱断言）。
func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		want []string
	}{
		{
			name: "default config",
			cfg:  config.DefaultConfig(),
			want: []string{
				"--bind", "127.0.0.1",
				"--port", "8080",
				"--log-level", "info",
				"--workers", "4",
				"--model", "qwen3-27b",
				"--ctx", "32768",
				"--threads", "0",
				"--quant", "q4_k_m",
				"--temp", "0.7",
				"--top-p", "0.9",
				"--top-k", "40",
				"--repeat", "1.1",
				"--max-tokens", "4096",
				"--cors", "*",
				"--timeout", "300",
				"--max-conns", "32",
				"--metrics",
				"--metrics-path", "/metrics",
				"--cache-dir", "./cache",
				"--cache-mb", "1024",
				"--evict", "lru",
			},
		},
		{
			name: "variant different port host model paths",
			cfg: config.Config{
				Host:           "0.0.0.0",
				Port:           9090,
				LogLevel:       "debug",
				Workers:        8,
				ModelName:      "qwen3-30b",
				CtxSize:        65536,
				Threads:        16,
				Quant:          "q8_0",
				Temperature:    1.0, // FormatFloat → "1"
				TopP:           0.5,
				TopK:           0,
				RepeatPenalty:  1.2,
				MaxTokens:      8192,
				CORSOrigins:    "http://a.example,http://b.example",
				TimeoutSec:     60,
				MaxConns:       128,
				MetricsEnabled: true,
				MetricsPath:    "/metrics",
				NumaNode:       config.DefNumaNode, // -1=自动，契约 ≥0 不满足，不输出 --numa
				Priority:       config.DefPriority, // ==normal，不输出 --priority
				CacheDir:       "D:/cache/ninfer",
				CacheSizeMB:    4096,
				EvictPolicy:    "fifo",
			},
			want: []string{
				"--bind", "0.0.0.0",
				"--port", "9090",
				"--log-level", "debug",
				"--workers", "8",
				"--model", "qwen3-30b",
				"--ctx", "65536",
				"--threads", "16",
				"--quant", "q8_0",
				"--temp", "1",
				"--top-p", "0.5",
				"--top-k", "0",
				"--repeat", "1.2",
				"--max-tokens", "8192",
				"--cors", "http://a.example,http://b.example",
				"--timeout", "60",
				"--max-conns", "128",
				"--metrics",
				"--metrics-path", "/metrics",
				"--cache-dir", "D:/cache/ninfer",
				"--cache-mb", "4096",
				"--evict", "fifo",
			},
		},
		{
			name: "all optional segments enabled",
			cfg: config.Config{
				Host:           "127.0.0.1",
				Port:           18080,
				LogLevel:       "warn",
				Workers:        2,
				ModelName:      "llama-8b",
				CtxSize:        8192,
				Threads:        4,
				Quant:          "f16",
				Temperature:    0.35,
				TopP:           0.95,
				TopK:           50,
				RepeatPenalty:  1.05,
				MaxTokens:      2048,
				APIKey:         "sk-live-12345",
				CORSOrigins:    "*",
				TimeoutSec:     120,
				MaxConns:       64,
				TLS:            true,
				CertPath:       "C:/certs/server.crt",
				KeyPath:        "C:/certs/server.key",
				MetricsEnabled: true,
				MetricsPath:    "/internal/metrics",
				LogFile:        "C:/logs/ninfer.log",
				MemLimitGB:     32,
				SwapGB:         8,
				NumaNode:       1,
				Priority:       "high",
				AffinityMask:   "0x0F",
				CacheDir:       "C:/data/cache",
				CacheSizeMB:    2048,
				EvictPolicy:    "lru",
				Prefetch:       true,
			},
			want: []string{
				"--bind", "127.0.0.1",
				"--port", "18080",
				"--log-level", "warn",
				"--workers", "2",
				"--model", "llama-8b",
				"--ctx", "8192",
				"--threads", "4",
				"--quant", "f16",
				"--temp", "0.35",
				"--top-p", "0.95",
				"--top-k", "50",
				"--repeat", "1.05",
				"--max-tokens", "2048",
				"--cors", "*",
				"--timeout", "120",
				"--max-conns", "64",
				"--api-key", "sk-live-12345",
				"--tls", "--cert", "C:/certs/server.crt", "--key", "C:/certs/server.key",
				"--metrics", "--metrics-path", "/internal/metrics",
				"--log-file", "C:/logs/ninfer.log",
				"--mem-limit-gb", "32",
				"--swap-gb", "8",
				"--numa", "1",
				"--priority", "high",
				"--affinity", "0x0F",
				"--cache-dir", "C:/data/cache",
				"--cache-mb", "2048",
				"--evict", "lru",
				"--prefetch",
			},
		},
		{
			name: "metrics disabled and float integer values",
			cfg: config.Config{
				Host:           "10.0.0.8",
				Port:           5000,
				LogLevel:       "error",
				Workers:        1,
				ModelName:      "tiny-model",
				CtxSize:        4096,
				Quant:          "q4_0",
				Temperature:    0, // FormatFloat → "0"
				TopP:           1, // FormatFloat → "1"
				TopK:           1,
				RepeatPenalty:  2, // FormatFloat → "2"
				MaxTokens:      1024,
				CORSOrigins:    "*",
				TimeoutSec:     10,
				MaxConns:       1,
				MetricsEnabled: false,
				MetricsPath:    "/metrics",
				NumaNode:       config.DefNumaNode, // -1=自动，契约 ≥0 不满足，不输出 --numa
				Priority:       config.DefPriority, // ==normal，不输出 --priority
				CacheDir:       "./cache",
				CacheSizeMB:    512,
				EvictPolicy:    "lru",
			},
			want: []string{
				"--bind", "10.0.0.8",
				"--port", "5000",
				"--log-level", "error",
				"--workers", "1",
				"--model", "tiny-model",
				"--ctx", "4096",
				"--threads", "0",
				"--quant", "q4_0",
				"--temp", "0",
				"--top-p", "1",
				"--top-k", "1",
				"--repeat", "2",
				"--max-tokens", "1024",
				"--cors", "*",
				"--timeout", "10",
				"--max-conns", "1",
				"--cache-dir", "./cache",
				"--cache-mb", "512",
				"--evict", "lru",
			},
		},
		{
			name: "metrics enabled but empty path",
			cfg: func() config.Config {
				cfg := config.DefaultConfig()
				cfg.MetricsPath = ""
				return cfg
			}(),
			want: []string{
				"--bind", "127.0.0.1",
				"--port", "8080",
				"--log-level", "info",
				"--workers", "4",
				"--model", "qwen3-27b",
				"--ctx", "32768",
				"--threads", "0",
				"--quant", "q4_k_m",
				"--temp", "0.7",
				"--top-p", "0.9",
				"--top-k", "40",
				"--repeat", "1.1",
				"--max-tokens", "4096",
				"--cors", "*",
				"--timeout", "300",
				"--max-conns", "32",
				"--metrics",
				"--cache-dir", "./cache",
				"--cache-mb", "1024",
				"--evict", "lru",
			},
		},
		{
			// 边界锁定：契约 NumaNode≥0（-1 才是自动哨兵）与 Priority≠normal 的字面语义。
			name: "numa zero boundary and empty priority",
			cfg: config.Config{
				Host:      "127.0.0.1",
				Port:      8080,
				LogLevel:  "info",
				Workers:   1,
				ModelName: "m",
				CtxSize:   1024,
				NumaNode:  0,
				Priority:  "",
			},
			want: []string{
				"--bind", "127.0.0.1",
				"--port", "8080",
				"--log-level", "info",
				"--workers", "1",
				"--model", "m",
				"--ctx", "1024",
				"--threads", "0",
				"--quant", "",
				"--temp", "0",
				"--top-p", "0",
				"--top-k", "0",
				"--repeat", "0",
				"--max-tokens", "0",
				"--cors", "",
				"--timeout", "0",
				"--max-conns", "0",
				"--numa", "0",
				"--priority", "",
				"--cache-dir", "",
				"--cache-mb", "0",
				"--evict", "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildArgs(tt.cfg)

			if len(got) != len(tt.want) {
				t.Fatalf("BuildArgs length = %d (got %s), want %d (want %s)",
					len(got), strings.Join(got, " "), len(tt.want), strings.Join(tt.want, " "))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("BuildArgs[%d] = %q, want %q\nfull got:  %s\nfull want: %s",
						i, got[i], tt.want[i], strings.Join(got, " "), strings.Join(tt.want, " "))
				}
			}
		})
	}
}
