package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Test_Validate_when_defaultConfig_returnsNoErrors
// 默认配置必须可启动：Validate(DefaultConfig()) 返回空列表。
func Test_Validate_when_defaultConfig_returnsNoErrors(t *testing.T) {
	if errs := Validate(DefaultConfig()); len(errs) != 0 {
		t.Fatalf("Validate(DefaultConfig()) = %v, want empty", errs)
	}
}

// Test_Validate_when_v1OutOfRange_returnsV1Errors
// V1：Port/Workers/CtxSize 合法范围。
func Test_Validate_when_v1OutOfRange_returnsV1Errors(t *testing.T) {
	tcs := []struct {
		name    string
		mutate  func(*Config)
		wantSub string
	}{
		{"portZero", func(c *Config) { c.Port = 0 }, "V1:Port=0"},
		{"portTooBig", func(c *Config) { c.Port = 65536 }, "V1:Port=65536"},
		{"workersZero", func(c *Config) { c.Workers = 0 }, "V1:Workers=0"},
		{"ctxSizeTooBig", func(c *Config) { c.CtxSize = 1048577 }, "V1:CtxSize=1048577"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			if errs := Validate(cfg); !errContains(errs, tc.wantSub) {
				t.Errorf("Validate = %v, want contains %q", errs, tc.wantSub)
			}
		})
	}
}

// Test_Validate_when_v2OutOfRange_returnsV2Errors
// V2：MaxTokens/TimeoutSec/MaxConns 下限 1。
func Test_Validate_when_v2OutOfRange_returnsV2Errors(t *testing.T) {
	tcs := []struct {
		name    string
		mutate  func(*Config)
		wantSub string
	}{
		{"maxTokensZero", func(c *Config) { c.MaxTokens = 0 }, "V2:MaxTokens=0"},
		{"timeoutSecZero", func(c *Config) { c.TimeoutSec = 0 }, "V2:TimeoutSec=0"},
		{"maxConnsZero", func(c *Config) { c.MaxConns = 0 }, "V2:MaxConns=0"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			if errs := Validate(cfg); !errContains(errs, tc.wantSub) {
				t.Errorf("Validate = %v, want contains %q", errs, tc.wantSub)
			}
		})
	}
}

// Test_Validate_when_v5LogLevelInvalid_returnsV5Error
// V5：LogLevel 必须为 debug/info/warn/error。
func Test_Validate_when_v5LogLevelInvalid_returnsV5Error(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LogLevel = "verbose"
	if errs := Validate(cfg); !errContains(errs, "V5:LogLevel=\"verbose\"") {
		t.Errorf("Validate = %v, want contains %q", errs, "V5:LogLevel=\"verbose\"")
	}
}

// Test_Validate_when_v3topPOutOfRange_returnsV3Error
// V3：TopP 必须位于 [0,1]。
func Test_Validate_when_v3topPOutOfRange_returnsV3Error(t *testing.T) {
	for _, v := range []float64{-0.1, 1.2} {
		cfg := DefaultConfig()
		cfg.TopP = v
		want := "V3:TopP=" + fmt.Sprintf("%v", v)
		if errs := Validate(cfg); !errContains(errs, want) {
			t.Errorf("Validate(TopP=%v) = %v, want contains %q", v, errs, want)
		}
	}
}

// Test_Validate_when_v5EvictPolicyInvalid_returnsV5Error
// V5：EvictPolicy 必须为 lru/fifo。
func Test_Validate_when_v5EvictPolicyInvalid_returnsV5Error(t *testing.T) {
	cfg := DefaultConfig()
	cfg.EvictPolicy = "random"
	if errs := Validate(cfg); !errContains(errs, `V5:EvictPolicy="random"`) {
		t.Errorf("Validate = %v, want contains %q", errs, `V5:EvictPolicy="random"`)
	}
}

// Test_Validate_when_v6ResourceInvalid_returnsV6Errors
// V6：MemLimitGB/SwapGB ≥ 0，NumaNode == -1 或 ≥ 0，CacheSizeMB ≥ 1。
func Test_Validate_when_v6ResourceInvalid_returnsV6Errors(t *testing.T) {
	tcs := []struct {
		name    string
		mutate  func(*Config)
		wantSub string
	}{
		{"memLimitNegative", func(c *Config) { c.MemLimitGB = -1 }, "V6:MemLimitGB=-1"},
		{"swapNegative", func(c *Config) { c.SwapGB = -1 }, "V6:SwapGB=-1"},
		{"cacheSizeZero", func(c *Config) { c.CacheSizeMB = 0 }, "V6:CacheSizeMB=0"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tc.mutate(&cfg)
			if errs := Validate(cfg); !errContains(errs, tc.wantSub) {
				t.Errorf("Validate = %v, want contains %q", errs, tc.wantSub)
			}
		})
	}
}

// Test_Validate_when_v7pathContainsNUL_returnsV7Error
// V7：路径类字段非空时不得含 NUL 字节。
func Test_Validate_when_v7pathContainsNUL_returnsV7Error(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ModelPath = "abc\x00def"
	if errs := Validate(cfg); !errContains(errs, "V7:") {
		t.Errorf("Validate = %v, want contains %q", errs, "V7:")
	}
}

// Test_Validate_when_multipleViolations_aggregatesAllErrors
// V8：不短路，一次返回全部错误。
func Test_Validate_when_multipleViolations_aggregatesAllErrors(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Port = 0
	cfg.TopP = 2
	cfg.EvictPolicy = "random"
	errs := Validate(cfg)
	for _, sub := range []string{"V1:Port=0", "V3:TopP=2", "V5:EvictPolicy=\"random\""} {
		if !errContains(errs, sub) {
			t.Errorf("Validate = %v, want contains %q", errs, sub)
		}
	}
}

// errContains 报告 errs 中是否存在包含 sub 的错误条目。
func errContains(errs []string, sub string) bool {
	for _, e := range errs {
		if contains(e, sub) {
			return true
		}
	}
	return false
}

// contains 是 strings.Contains 的别名，避免测试文件与实现文件 import 名冲突。
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (indexOf(s, sub) >= 0)
}

// indexOf 返回 sub 在 s 中的首个下标，未找到返回 -1。
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// Test_Validate_when_v6fileExistsNoError 占位说明：V4/V6 的文件存在性检查
// 使用 os.Stat，此处通过真实临时文件验证「存在 → 无 V4 错误」。
func Test_Validate_when_v4fileExistsNoV4Error(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing-model.bin")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("creating temp model file: %v", err)
	}
	// ModelPath 指向已存在文件，不应产生任何 V4 文件存在性错误。
	cfg := DefaultConfig()
	cfg.ModelPath = path
	if errs := Validate(cfg); errContains(errs, "V4:") {
		t.Errorf("Validate with existing ModelPath = %v, want no V4 errors", errs)
	}
}
