package config

import (
	"strings"
	"testing"
)

// validCfg 返回通过校验的基础配置（模型文件指向临时目录真实文件）。
func validCfg(t *testing.T) Config {
	t.Helper()
	cfg := DefaultConfig()
	cfg.ModelPath = newModelFile(t)
	return cfg
}

// Test_Validate_when_batRecipe_returnsNoErrors
// bat 默认配方（模型文件存在）必须通过校验。
func Test_Validate_when_batRecipe_returnsNoErrors(t *testing.T) {
	if errs := Validate(validCfg(t)); len(errs) != 0 {
		t.Fatalf("Validate(bat recipe) = %v, want empty", errs)
	}
}

// Test_Validate_when_modelMissing_returnsError 模型文件不存在/为目录 → 报错。
func Test_Validate_when_modelMissing_returnsError(t *testing.T) {
	cfg := validCfg(t)
	cfg.ModelPath = `Z:\no\such\model.ninfer`
	if errs := Validate(cfg); len(errs) == 0 || !strings.Contains(errs[0], "模型文件不存在") {
		t.Fatalf("Validate(missing model) = %v, want 模型文件不存在", errs)
	}
}

// Test_Validate_when_outOfRange_returnsErrors 基本范围校验。
func Test_Validate_when_outOfRange_returnsErrors(t *testing.T) {
	tcs := []struct {
		name    string
		mutate  func(*Config)
		wantSub string
	}{
		{"portZero", func(c *Config) { c.Port = 0 }, "监听端口"},
		{"portTooBig", func(c *Config) { c.Port = 65536 }, "监听端口"},
		{"concurrencyZero", func(c *Config) { c.Concurrency = 0 }, "并发解码请求数"},
		{"concurrencyNine", func(c *Config) { c.Concurrency = 9 }, "并发解码请求数"},
		{"contextSmall", func(c *Config) { c.Context = 255 }, "上下文上限"},
		{"budgetSmall", func(c *Config) { c.Budget = 127 }, "历史预算"},
		{"reserveSmall", func(c *Config) { c.Reserve = 63 }, "生成保留"},
		{"hostMiBZero", func(c *Config) { c.HostMiB = 0 }, "Host KV 预算"},
		{"maxTokensZero", func(c *Config) { c.MaxTokens = 0 }, "默认输出上限"},
		{"prefillSmall", func(c *Config) { c.Prefill = 127 }, "prefill 块"},
		{"sessionsZero", func(c *Config) { c.RetainedSessions = 0 }, "保留历史数"},
		{"sessions17", func(c *Config) { c.RetainedSessions = 17 }, "保留历史数"},
		{"mtp16", func(c *Config) { c.MtpDrafts = 16 }, "MTP 草稿数"},
		{"ngram64", func(c *Config) { c.NgramDrafts = 64 }, "ngram 草稿数"},
		{"minMatch3", func(c *Config) { c.NgramMinMatch = 3 }, "ngram 最短匹配"},
		{"visionTokens0", func(c *Config) { c.VisionTokens = 0 }, "视觉合并上限"},
		{"diskMiB0", func(c *Config) { c.DiskMiB = 0 }, "SSD 配额"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validCfg(t)
			tc.mutate(&cfg)
			errs := Validate(cfg)
			if len(errs) == 0 {
				t.Fatalf("Validate(%s) = empty, want error containing %q", tc.name, tc.wantSub)
			}
			found := false
			for _, e := range errs {
				if strings.Contains(e, tc.wantSub) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Validate(%s) = %v, want error containing %q", tc.name, errs, tc.wantSub)
			}
		})
	}
}

// Test_Validate_when_alignmentViolated_returnsErrors 对齐与组合约束。
func Test_Validate_when_alignmentViolated_returnsErrors(t *testing.T) {
	tcs := []struct {
		name    string
		mutate  func(*Config)
		wantSub string
	}{
		{"budgetNotAligned64", func(c *Config) { c.Budget = 36865 }, "对齐 64"},
		{"reserveNotAligned64", func(c *Config) { c.Reserve = 16385 }, "对齐 64"},
		{"prefillNotAligned128", func(c *Config) { c.Prefill = 513 }, "对齐 128"},
		{"prefillExceedsReserve", func(c *Config) { c.Prefill = 16512 }, "不得超过生成保留"},
		{"budgetPlusReserveExceedsContext", func(c *Config) { c.Context = 4096 }, "不得超过上下文上限"},
		{"recentTokens63", func(c *Config) { c.RecentTokens = 63 }, "必须 ≥ 64"},
		{"sinkRecentExceedBudget", func(c *Config) { c.Budget = 128 }, "超过历史预算"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validCfg(t)
			tc.mutate(&cfg)
			errs := Validate(cfg)
			found := false
			for _, e := range errs {
				if strings.Contains(e, tc.wantSub) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Validate(%s) = %v, want error containing %q", tc.name, errs, tc.wantSub)
			}
		})
	}
}

// Test_Validate_when_enumInvalid_returnsErrors 枚举校验。
func Test_Validate_when_enumInvalid_returnsErrors(t *testing.T) {
	tcs := []struct {
		name    string
		mutate  func(*Config)
		wantSub string
	}{
		{"kvType", func(c *Config) { c.KvType = "q4_0" }, "KV 存储"},
		{"visionResidency", func(c *Config) { c.VisionResidency = "disk" }, "视觉驻留"},
		{"deviceProfile", func(c *Config) { c.DeviceProfile = "calibrate" }, "设备档案"},
		{"scoreAgg", func(c *Config) { c.ScoreAggregation = "sum" }, "评分聚合"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validCfg(t)
			tc.mutate(&cfg)
			errs := Validate(cfg)
			found := false
			for _, e := range errs {
				if strings.Contains(e, tc.wantSub) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Validate(%s) = %v, want error containing %q", tc.name, errs, tc.wantSub)
			}
		})
	}
}

// Test_Validate_when_combinationViolated_returnsErrors 组合约束。
func Test_Validate_when_combinationViolated_returnsErrors(t *testing.T) {
	tcs := []struct {
		name    string
		mutate  func(*Config)
		wantSub string
	}{
		{"adaptiveWithoutMtp", func(c *Config) { c.MtpDrafts = 0 }, "自适应 MTP"},
		{"fastPrefillWithBf16", func(c *Config) { c.KvType = "bf16"; c.FastPrefill = true }, "快速 prefill"},
		{"cpuVisionWithoutVision", func(c *Config) { c.Vision = false }, "CPU 视觉驻留"},
		{"ngramWithoutMtp", func(c *Config) { c.MtpDrafts = 0 }, "ngram 草稿需要"},
		{"badReasoningEffort", func(c *Config) { c.DefaultReasoningEffort = "ultra" }, "思考力度"},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validCfg(t)
			tc.mutate(&cfg)
			errs := Validate(cfg)
			found := false
			for _, e := range errs {
				if strings.Contains(e, tc.wantSub) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Validate(%s) = %v, want error containing %q", tc.name, errs, tc.wantSub)
			}
		})
	}
}

// Test_Validate_when_gpuInvalid_returnsErrors GPU 校验。
func Test_Validate_when_gpuInvalid_returnsErrors(t *testing.T) {
	tcs := []struct {
		name   string
		mutate func(*Config)
	}{
		{"empty", func(c *Config) { c.Gpu = "" }},
		{"negative", func(c *Config) { c.Gpu = "-1" }},
		{"dup", func(c *Config) { c.Gpu = "0,0" }},
		{"tooMany", func(c *Config) { c.Gpu = "0,1,2,3,4,5,6,7,8" }},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validCfg(t)
			tc.mutate(&cfg)
			if errs := Validate(cfg); len(errs) == 0 {
				t.Fatalf("Validate(%s) = empty, want GPU error", tc.name)
			}
		})
	}
}

// Test_Validate_when_stageLayersInvalid_returnsErrors 多卡层数校验。
func Test_Validate_when_stageLayersInvalid_returnsErrors(t *testing.T) {
	tcs := []struct {
		name   string
		mutate func(*Config)
	}{
		{"singleCard", func(c *Config) { c.StageLayers = "24" }},
		{"zeroLayer", func(c *Config) { c.Gpu = "0,1"; c.StageLayers = "24,0" }},
		{"notNumber", func(c *Config) { c.Gpu = "0,1"; c.StageLayers = "24,x" }},
		{"moreGpusNeeded", func(c *Config) { c.StageLayers = "24,24" }},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validCfg(t)
			tc.mutate(&cfg)
			if errs := Validate(cfg); len(errs) == 0 {
				t.Fatalf("Validate(%s) = empty, want stage-layers error", tc.name)
			}
		})
	}
}

// Test_Validate_when_stageLayersValid_multiGpu 合法多卡配置通过。
func Test_Validate_when_stageLayersValid_multiGpu(t *testing.T) {
	cfg := validCfg(t)
	cfg.Gpu = "0,1"
	cfg.StageLayers = "24,24"
	if errs := Validate(cfg); len(errs) != 0 {
		t.Fatalf("Validate(multi-gpu) = %v, want empty", errs)
	}
}
