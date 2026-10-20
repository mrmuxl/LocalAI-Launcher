package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Test_SaveLlamaConfig_when_roundTrip_preservesAll46Fields
// 写入临时目录 → LoadLlamaConfig 读回 → 整结构相等（JSON 往返）。
func Test_SaveLlamaConfig_when_roundTrip_preservesAll46Fields(t *testing.T) {
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
	if cfg != DefaultLlamaConfig() {
		t.Error("missing file should yield defaults")
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
