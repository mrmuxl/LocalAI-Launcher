package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Test_Validate_when_v1Boundaries_allAccepted
// V1 合法边界值：Port/Workers/CtxSize 各取下界 1 与上界 65535/256/1048576 → 无错误。
func Test_Validate_when_v1Boundaries_allAccepted(t *testing.T) {
	tcs := []struct {
		name   string
		mutate func(*Config)
	}{
		{"portLower1", func(c *Config) { c.Port = 1 }},
		{"portUpper65535", func(c *Config) { c.Port = 65535 }},
		{"workersLower1", func(c *Config) { c.Workers = 1 }},
		{"workersUpper256", func(c *Config) { c.Workers = 256 }},
		{"ctxSizeLower1", func(c *Config) { c.CtxSize = 1 }},
		{"ctxSizeUpper1048576", func(c *Config) { c.CtxSize = 1048576 }},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			if errs := Validate(cfg); len(errs) != 0 {
				t.Errorf("Validate(%s) = %v, want no error", tc.name, errs)
			}
		})
	}
}

// Test_Validate_when_v1BoundariesCrossed_rejected
// V1 越界反例：Port/Workers/CtxSize 下界-1（0）与上界+1 → 对应 V1 错误。
func Test_Validate_when_v1BoundariesCrossed_rejected(t *testing.T) {
	tcs := []struct {
		name    string
		mutate  func(*Config)
		wantSub string
	}{
		{"portZero", func(c *Config) { c.Port = 0 }, "V1:Port=0"},
		{"port65536", func(c *Config) { c.Port = 65536 }, "V1:Port=65536"},
		{"workersZero", func(c *Config) { c.Workers = 0 }, "V1:Workers=0"},
		{"workers257", func(c *Config) { c.Workers = 257 }, "V1:Workers=257"},
		{"ctxSizeZero", func(c *Config) { c.CtxSize = 0 }, "V1:CtxSize=0"},
		{"ctxSize1048577", func(c *Config) { c.CtxSize = 1048577 }, "V1:CtxSize=1048577"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			if errs := Validate(cfg); !errContains(errs, tc.wantSub) {
				t.Errorf("Validate(%s) = %v, want contains %q", tc.name, errs, tc.wantSub)
			}
		})
	}
}

// Test_Validate_when_v2Boundaries_allAccepted
// V2 合法边界值：MaxTokens/TimeoutSec 下界 1；MaxConns 下界 1 上界 1024 → 无错误。
func Test_Validate_when_v2Boundaries_allAccepted(t *testing.T) {
	tcs := []struct {
		name   string
		mutate func(*Config)
	}{
		{"maxTokens1", func(c *Config) { c.MaxTokens = 1 }},
		{"timeoutSec1", func(c *Config) { c.TimeoutSec = 1 }},
		{"maxConns1", func(c *Config) { c.MaxConns = 1 }},
		{"maxConns1024", func(c *Config) { c.MaxConns = 1024 }},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			if errs := Validate(cfg); len(errs) != 0 {
				t.Errorf("Validate(%s) = %v, want no error", tc.name, errs)
			}
		})
	}
}

// Test_Validate_when_v2BoundariesCrossed_rejected
// V2 越界反例：MaxTokens/TimeoutSec 为 0；MaxConns 上界+1（1025）→ 对应 V2 错误。
func Test_Validate_when_v2BoundariesCrossed_rejected(t *testing.T) {
	tcs := []struct {
		name    string
		mutate  func(*Config)
		wantSub string
	}{
		{"maxTokensZero", func(c *Config) { c.MaxTokens = 0 }, "V2:MaxTokens=0"},
		{"timeoutSecZero", func(c *Config) { c.TimeoutSec = 0 }, "V2:TimeoutSec=0"},
		{"maxConnsZero", func(c *Config) { c.MaxConns = 0 }, "V2:MaxConns=0"},
		{"maxConns1025", func(c *Config) { c.MaxConns = 1025 }, "V2:MaxConns=1025"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			if errs := Validate(cfg); !errContains(errs, tc.wantSub) {
				t.Errorf("Validate(%s) = %v, want contains %q", tc.name, errs, tc.wantSub)
			}
		})
	}
}

// Test_Validate_when_v3Boundaries_allAccepted
// V3 合法边界值：Temperature/TopP/TopK 下界 0；TopP 上界 1；RepeatPenalty 正值 → 无错误。
func Test_Validate_when_v3Boundaries_allAccepted(t *testing.T) {
	tcs := []struct {
		name   string
		mutate func(*Config)
	}{
		{"temperatureZero", func(c *Config) { c.Temperature = 0 }},
		{"topPLower0", func(c *Config) { c.TopP = 0 }},
		{"topPUpper1", func(c *Config) { c.TopP = 1 }},
		{"topKZero", func(c *Config) { c.TopK = 0 }},
		{"repeatPenalty1", func(c *Config) { c.RepeatPenalty = 1 }},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			if errs := Validate(cfg); len(errs) != 0 {
				t.Errorf("Validate(%s) = %v, want no error", tc.name, errs)
			}
		})
	}
}

// Test_Validate_when_v3BoundariesCrossed_rejected
// V3 越界反例：Temperature 负值；TopP 越界两侧；TopK 负值；RepeatPenalty 0（<=0）→ 对应 V3 错误。
func Test_Validate_when_v3BoundariesCrossed_rejected(t *testing.T) {
	tcs := []struct {
		name    string
		mutate  func(*Config)
		wantSub string
	}{
		{"temperatureNegative", func(c *Config) { c.Temperature = -0.5 }, "V3:Temperature=-0.5"},
		{"topPBelow0", func(c *Config) { c.TopP = -0.1 }, "V3:TopP=-0.1"},
		{"topPAbove1", func(c *Config) { c.TopP = 1.5 }, "V3:TopP=1.5"},
		{"topKNegative", func(c *Config) { c.TopK = -1 }, "V3:TopK=-1"},
		{"repeatPenaltyZero", func(c *Config) { c.RepeatPenalty = 0 }, "V3:RepeatPenalty=0"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			if errs := Validate(cfg); !errContains(errs, tc.wantSub) {
				t.Errorf("Validate(%s) = %v, want contains %q", tc.name, errs, tc.wantSub)
			}
		})
	}
}

// Test_Validate_when_v4TlsEnabled_certAndKeyFileChecks
// V4：TLS 开启时证书/私钥路径非空且文件必须存在（t.TempDir 真实文件）。
func Test_Validate_when_v4TlsEnabled_certAndKeyFileChecks(t *testing.T) {
	dir := t.TempDir()
	cert := filepath.Join(dir, "cert.pem")
	key := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(cert, []byte("cert"), 0o600); err != nil {
		t.Fatalf("创建证书临时文件: %v", err)
	}
	if err := os.WriteFile(key, []byte("key"), 0o600); err != nil {
		t.Fatalf("创建私钥临时文件: %v", err)
	}

	t.Run("合法边界：两文件均存在 → 无 V4 错误", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.TLS = true
		cfg.CertPath = cert
		cfg.KeyPath = key
		if errs := Validate(cfg); len(errs) != 0 {
			t.Fatalf("Validate(TLS+真实证书/私钥) = %v, want 无错误", errs)
		}
	})

	t.Run("反例：证书文件缺失 → V4 错误", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.TLS = true
		cfg.CertPath = filepath.Join(dir, "not-exist-cert.pem")
		cfg.KeyPath = key
		errs := Validate(cfg)
		if len(errs) == 0 {
			t.Fatal("Validate(TLS+缺失证书) = 无错误, want V4 错误")
		}
		if !hasV4Error(errs) {
			t.Fatalf("Validate(TLS+缺失证书) = %v, want 含 V4 前缀错误", errs)
		}
	})

	t.Run("反例：私钥文件缺失 → V4 错误", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.TLS = true
		cfg.CertPath = cert
		cfg.KeyPath = filepath.Join(dir, "not-exist-key.pem")
		errs := Validate(cfg)
		if len(errs) == 0 {
			t.Fatal("Validate(TLS+缺失私钥) = 无错误, want V4 错误")
		}
		if !hasV4Error(errs) {
			t.Fatalf("Validate(TLS+缺失私钥) = %v, want 含 V4 前缀错误", errs)
		}
	})

	t.Run("反例：证书与私钥均缺失 → 两条 V4 错误（聚合）", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.TLS = true
		cfg.CertPath = filepath.Join(dir, "missing-a.pem")
		cfg.KeyPath = filepath.Join(dir, "missing-b.pem")
		errs := Validate(cfg)
		if countV4Errors(errs) != 2 {
			t.Fatalf("Validate(TLS+双缺失) = %v, want 恰好 2 条 V4 错误", errs)
		}
	})

	t.Run("合法边界：TLS 关闭时不检查证书文件", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.TLS = false
		cfg.CertPath = filepath.Join(dir, "missing-cert.pem")
		cfg.KeyPath = filepath.Join(dir, "missing-key.pem")
		if errs := Validate(cfg); len(errs) != 0 {
			t.Fatalf("Validate(TLS 关闭+缺失文件) = %v, want 无错误", errs)
		}
	})
}

// Test_Validate_when_v5enumValues_allAccepted
// V5 合法边界：LogLevel/Priority/EvictPolicy 取全部合法枚举值 → 无错误。
func Test_Validate_when_v5enumValues_allAccepted(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error"} {
		cfg := DefaultConfig()
		cfg.LogLevel = level
		if errs := Validate(cfg); len(errs) != 0 {
			t.Fatalf("Validate(LogLevel=%q) = %v, want 无错误", level, errs)
		}
	}
	for _, prio := range []string{"low", "normal", "high", "realtime"} {
		cfg := DefaultConfig()
		cfg.Priority = prio
		if errs := Validate(cfg); len(errs) != 0 {
			t.Fatalf("Validate(Priority=%q) = %v, want 无错误", prio, errs)
		}
	}
	for _, ev := range []string{"lru", "fifo"} {
		cfg := DefaultConfig()
		cfg.EvictPolicy = ev
		if errs := Validate(cfg); len(errs) != 0 {
			t.Fatalf("Validate(EvictPolicy=%q) = %v, want 无错误", ev, errs)
		}
	}
}

// Test_Validate_when_v5enumValues_outOfSet_rejected
// V5 越界反例：三个枚举字段各取一个集合外值 → 均产生对应 V5 错误。
func Test_Validate_when_v5enumValues_outOfSet_rejected(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*Config)
		wantSub string
	}{
		{"LogLevel=verbose", func(c *Config) { c.LogLevel = "verbose" }, `V5:LogLevel="verbose"`},
		{"Priority=urgent", func(c *Config) { c.Priority = "urgent" }, `V5:Priority="urgent"`},
		{"EvictPolicy=lfu", func(c *Config) { c.EvictPolicy = "lfu" }, `V5:EvictPolicy="lfu"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			errs := Validate(cfg)
			if len(errs) == 0 {
				t.Fatalf("Validate(%s) = 无错误, want V5 错误", tc.name)
			}
			if !errContains(errs, tc.wantSub) {
				t.Fatalf("Validate(%s) = %v, want 含 %s", tc.name, errs, tc.wantSub)
			}
		})
	}
}

// Test_Validate_when_v6portConflict_outOfScopeForV1
// V6 说明：计划 V6「端口冲突」标注 v1 不做（validate.go 文档注释确认），
// 本规则由后续 T5 进程层负责；此处仅断言 V6 不在 config 包 Validate 中误报：
// 同一端口配置两次校验结果一致且不含端口冲突字样。
func Test_Validate_when_v6portConflict_outOfScopeForV1(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Port = 18080
	first := Validate(cfg)
	second := Validate(cfg)
	if len(first) != 0 {
		t.Fatalf("Validate(Port=18080 合法) = %v, want 无错误", first)
	}
	if len(second) != 0 || !reflect.DeepEqual(first, second) {
		t.Fatalf("V6 稳定性检查失败: first=%v second=%v", first, second)
	}
}

// Test_Validate_when_v6ResourceBoundaries
// V6 数值边界：MemLimitGB/SwapGB 下界 0；NumaNode 须为 -1 或 >= 0；CacheSizeMB 下界 1。
func Test_Validate_when_v6ResourceBoundaries(t *testing.T) {
	positive := []struct {
		name   string
		mutate func(*Config)
	}{
		{"memLimitGB0", func(c *Config) { c.MemLimitGB = 0 }},
		{"swapGB0", func(c *Config) { c.SwapGB = 0 }},
		{"numaNodeMinus1", func(c *Config) { c.NumaNode = -1 }},
		{"numaNode0", func(c *Config) { c.NumaNode = 0 }},
		{"cacheSizeMB1", func(c *Config) { c.CacheSizeMB = 1 }},
	}
	for _, tc := range positive {
		t.Run("正例："+tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			if errs := Validate(cfg); errContains(errs, "V6:") {
				t.Errorf("Validate(%s) = %v, want 无 V6 错误", tc.name, errs)
			}
		})
	}
	negative := []struct {
		name    string
		mutate  func(*Config)
		wantSub string
	}{
		{"memLimitGBMinus1", func(c *Config) { c.MemLimitGB = -1 }, "V6:MemLimitGB=-1 < 0"},
		{"swapGBMinus1", func(c *Config) { c.SwapGB = -1 }, "V6:SwapGB=-1 < 0"},
		{"numaNodeMinus2", func(c *Config) { c.NumaNode = -2 }, "V6:NumaNode=-2 must be -1 or >= 0"},
		{"cacheSizeMB0", func(c *Config) { c.CacheSizeMB = 0 }, "V6:CacheSizeMB=0 < 1"},
	}
	for _, tc := range negative {
		t.Run("反例："+tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			if errs := Validate(cfg); !errContains(errs, tc.wantSub) {
				t.Errorf("Validate(%s) = %v, want 含 %q", tc.name, errs, tc.wantSub)
			}
		})
	}
}

// Test_Validate_when_v7pathFields_nulRejectedEmptyAllowed
// V7 合法边界：5 个路径字段（ModelPath/CacheDir/LogFile/CertPath/KeyPath）
// 空字符串不报错；反例：任一路径含 NUL 字节均报 V7 错误。
func Test_Validate_when_v7pathFields_nulRejectedEmptyAllowed(t *testing.T) {
	t.Run("合法边界：路径全空 → 无 V7 错误", func(t *testing.T) {
		cfg := Config{}
		if hasV7Error(Validate(cfg)) {
			t.Fatalf("Validate(零值配置) = %v, want 无 V7 错误（路径全空）", Validate(cfg))
		}
	})

	pathCases := []struct {
		name   string
		mutate func(*Config)
	}{
		{"ModelPath", func(c *Config) { c.ModelPath = "a\x00b" }},
		{"CacheDir", func(c *Config) { c.CacheDir = "x\x00" }},
		{"LogFile", func(c *Config) { c.LogFile = "\x00log" }},
		{"CertPath", func(c *Config) { c.CertPath = "c\x00ert" }},
		{"KeyPath", func(c *Config) { c.KeyPath = "k\x00ey" }},
	}
	for _, tc := range pathCases {
		t.Run("反例："+tc.name+" 含 NUL → V7 错误", func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			errs := Validate(cfg)
			if !hasV7Error(errs) {
				t.Fatalf("Validate(%s 含 NUL) = %v, want V7 错误", tc.name, errs)
			}
		})
	}
}

// Test_Validate_when_v8aggregate_stableOrderAndExactCount
// V8：聚合返回不短路——违反全部可触发规则的同一输入：
// 两次调用结果逐条一致（顺序稳定），且错误总数恰好等于预期 9 条（不多不少）。
func Test_Validate_when_v8aggregate_stableOrderAndExactCount(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Port = 0              // V1
	cfg.Workers = 0           // V1
	cfg.MaxTokens = -1        // V2
	cfg.LogLevel = "bogus"    // V5
	cfg.Priority = "bogus"    // V5
	cfg.EvictPolicy = "bogus" // V5
	cfg.TLS = true            // V4 x2（证书+私钥均缺失）
	cfg.CertPath = "missing-cert.pem"
	cfg.KeyPath = "missing-key.pem"
	cfg.ModelPath = "m\x00" // V7

	first := Validate(cfg)
	second := Validate(cfg)
	if !sliceEqual(first, second) {
		t.Fatalf("V8 顺序稳定性失败：同一输入两次结果不一致\nfirst=%v\nsecond=%v", first, second)
	}
	if len(first) != 9 {
		t.Fatalf("V8 聚合完整性失败：%d 条错误, want 恰好 9 条: %v", len(first), first)
	}
	expected := []string{
		`V1:Port=0 out of range [1,65535]`,
		`V1:Workers=0 out of range [1,256]`,
		`V2:MaxTokens=-1 < 1`,
		`V4:cert file not found: missing-cert.pem`,
		`V4:key file not found: missing-key.pem`,
		`V5:LogLevel="bogus" not in {debug,info,warn,error}`,
		`V5:Priority="bogus" not in {low,normal,high,realtime}`,
		`V5:EvictPolicy="bogus" not in {lru,fifo}`,
		`V7:path contains NUL byte`,
	}
	// 精确总数 9 已断言；此处逐条核对关键错误码前缀（V1x2、V2x1、V4x2、V5x3、V7x1）。
	for _, want := range expected {
		if !errContains(first, want) {
			t.Errorf("V8 聚合缺少预期错误 %q，实际: %v", want, first)
		}
	}
}

// ── 本地 helper（不与 validate_test.go 的 errContains 冲突）──────────────

func hasV4Error(errs []string) bool { return errContains(errs, "V4:") }
func hasV7Error(errs []string) bool { return errContains(errs, "V7:") }

func countV4Errors(errs []string) int {
	n := 0
	for _, e := range errs {
		if errContains([]string{e}, "V4:") {
			n++
		}
	}
	return n
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
