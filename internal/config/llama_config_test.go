package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Test_SaveLlamaConfig_when_roundTrip_preservesAll52Fields
// 写入临时目录 → LoadLlamaConfig 读回 → 整结构相等（JSON 往返）。
func Test_SaveLlamaConfig_when_roundTrip_preservesAll52Fields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "llama.json")
	cfg := DefaultLlamaConfig()
	cfg.Host = "127.0.0.1"
	cfg.Port = 9090
	cfg.CacheTypeV = "q8_0"
	cfg.FlashAttn = "off"
	cfg.Jinja = false

	if err := SaveLlamaConfig(cfg, path); err != nil {
		t.Fatalf("SaveLlamaConfig: %v", err)
	}

	got, err := LoadLlamaConfig(path)
	if err != nil {
		t.Fatalf("LoadLlamaConfig after SaveLlamaConfig: %v", err)
	}
	if got != cfg {
		t.Errorf("round trip mismatch:\ngot:  %+v\nwant: %+v", got, cfg)
	}
}

// Test_LoadLlamaConfig_when_partialJSON_mergesOntoDefaults
// 文件存在但字段缺失：以默认值为底，文件中出现的字段覆盖对应默认值。
func Test_LoadLlamaConfig_when_partialJSON_mergesOntoDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "llama.json")
	partial := `{"port":9090,"threads":8,"cacheTypeV":"q8_0"}`
	if err := os.WriteFile(path, []byte(partial), 0o644); err != nil {
		t.Fatalf("write partial file: %v", err)
	}

	cfg, err := LoadLlamaConfig(path)

	if err != nil {
		t.Fatalf("LoadLlamaConfig(partial) unexpected error: %v", err)
	}
	if cfg.Port != 9090 || cfg.Threads != 8 || cfg.CacheTypeV != "q8_0" {
		t.Errorf("overridden fields wrong: Port=%d Threads=%d CacheTypeV=%q",
			cfg.Port, cfg.Threads, cfg.CacheTypeV)
	}
	if cfg.Host != "0.0.0.0" || cfg.CtxSize != 81920 || cfg.GpuLayers != 99 {
		t.Errorf("untouched fields not defaulted: Host=%q CtxSize=%d GpuLayers=%d",
			cfg.Host, cfg.CtxSize, cfg.GpuLayers)
	}
}

// Test_LoadLlamaConfig_when_missing_returnsDefaults
// 文件不存在 → 返回默认值且不报错。
func Test_LoadLlamaConfig_when_missing_returnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent", "llama.json")

	cfg, err := LoadLlamaConfig(path)
	if err != nil {
		t.Fatalf("LoadLlamaConfig(missing) unexpected error: %v", err)
	}
	want := DefaultLlamaConfig()
	want.ModelPath = "" // 缺文件时模型/视觉/模板路径不猜默认，为空等用户指定
	want.MmprojPath = ""
	want.ChatTemplateFile = ""
	if cfg != want {
		t.Error("missing file should yield defaults with empty path fields")
	}
}

// Test_LoadLlamaConfig_when_badJSON_returnsError
// JSON 损坏 → 返回错误且不写盘。
func Test_LoadLlamaConfig_when_badJSON_returnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "llama.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write bad file: %v", err)
	}

	if _, err := LoadLlamaConfig(path); err == nil {
		t.Fatal("LoadLlamaConfig(bad json) expected error, got nil")
	}
}

// moeValidateBase 构建 ValidateLlama MoE 用例配置：默认值 + 有效模型文件，
// mutate 调整 MoE 字段，返回校验错误列表。
func moeValidateBase(t *testing.T, mutate func(*LlamaConfig)) []string {
	t.Helper()
	model := filepath.Join(t.TempDir(), "model.gguf")
	if err := os.WriteFile(model, []byte("dummy"), 0o644); err != nil {
		t.Fatalf("write dummy model: %v", err)
	}
	cfg := DefaultLlamaConfig()
	cfg.ModelPath = model
	mutate(&cfg)
	return ValidateLlama(cfg)
}

// containsErr 判断 errs 是否含指定文案。
func containsErr(errs []string, want string) bool {
	for _, e := range errs {
		if e == want {
			return true
		}
	}
	return false
}

// Test_LoadLlamaConfig_when_oldJSONwithoutMoEFields_appliesMoEDefaults
// 旧版 JSON（无 MoE 字段）→ MoE 字段取默认值（dense/off/0/0），向后兼容。
func Test_LoadLlamaConfig_when_oldJSONwithoutMoEFields_appliesMoEDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "llama.json")
	old := `{"port":9090,"gpuLayers":99}`
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatalf("write old file: %v", err)
	}

	cfg, err := LoadLlamaConfig(path)
	if err != nil {
		t.Fatalf("LoadLlamaConfig(old json) unexpected error: %v", err)
	}
	if cfg.MoeModelType != "dense" || cfg.MoeCpuMode != "off" ||
		cfg.NCpuMoe != 0 || cfg.MoeCacheMB != 0 {
		t.Errorf("MoE defaults not applied: MoeModelType=%q MoeCpuMode=%q NCpuMoe=%d MoeCacheMB=%d",
			cfg.MoeModelType, cfg.MoeCpuMode, cfg.NCpuMoe, cfg.MoeCacheMB)
	}
}

// Test_ValidateLlama_when_moeFirstN_nonPositiveN_reportsError
// moe + firstN + N=0 → 报「必须大于 0」；N=-1 → 同时报「不能为负」。
func Test_ValidateLlama_when_moeFirstN_nonPositiveN_reportsError(t *testing.T) {
	errs := moeValidateBase(t, func(c *LlamaConfig) {
		c.MoeModelType = "moe"
		c.MoeCpuMode = "firstN"
		c.NCpuMoe = 0
	})
	if !containsErr(errs, "驻留 CPU 层数（--n-cpu-moe）在前 N 层模式下必须大于 0") {
		t.Fatalf("N=0 应报「必须大于 0」, errs=%q", errs)
	}

	errs = moeValidateBase(t, func(c *LlamaConfig) {
		c.MoeModelType = "moe"
		c.MoeCpuMode = "firstN"
		c.NCpuMoe = -1
	})
	if !containsErr(errs, "驻留 CPU 层数（--n-cpu-moe）在前 N 层模式下必须大于 0") ||
		!containsErr(errs, "驻留 CPU 层数（--n-cpu-moe）不能为负") {
		t.Fatalf("N=-1 应同时报「必须大于 0」和「不能为负」, errs=%q", errs)
	}
}

// Test_ValidateLlama_when_moe_negativeMoeCacheMB_reportsError
// moe + MoeCacheMB=-1 → 报「MoE 专家缓存（MiB）不能为负」。
func Test_ValidateLlama_when_moe_negativeMoeCacheMB_reportsError(t *testing.T) {
	errs := moeValidateBase(t, func(c *LlamaConfig) {
		c.MoeModelType = "moe"
		c.MoeCacheMB = -1
	})
	if !containsErr(errs, "MoE 专家缓存（MiB）不能为负") {
		t.Fatalf("MoeCacheMB=-1 应报「不能为负」, errs=%q", errs)
	}
}

// Test_ValidateLlama_when_moe_invalidMoeCpuMode_reportsError
// moe + MoeCpuMode="bogus" → 报「必须为 off/all/firstN」。
func Test_ValidateLlama_when_moe_invalidMoeCpuMode_reportsError(t *testing.T) {
	errs := moeValidateBase(t, func(c *LlamaConfig) {
		c.MoeModelType = "moe"
		c.MoeCpuMode = "bogus"
	})
	if !containsErr(errs, "MoE CPU 卸载模式（--cpu-moe）必须为 off/all/firstN") {
		t.Fatalf("非法 MoeCpuMode 应报「必须为 off/all/firstN」, errs=%q", errs)
	}
}

// Test_ValidateLlama_when_dense_anyMoEValues_noError
// dense 忽略全部 MoE 字段：bogus 模式 + 负值也不产生 MoE 错误。
func Test_ValidateLlama_when_dense_anyMoEValues_noError(t *testing.T) {
	errs := moeValidateBase(t, func(c *LlamaConfig) {
		c.MoeModelType = "dense"
		c.MoeCpuMode = "bogus"
		c.NCpuMoe = -5
		c.MoeCacheMB = -9
	})
	if len(errs) != 0 {
		t.Fatalf("dense 应忽略 MoE 字段, errs=%q", errs)
	}
}
