package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Test_LoadConfig_when_fileMissing_returnsDefaultsNoError
// load.go 文档契约：文件不存在 → 返回 DefaultConfig()，不报错。
func Test_LoadConfig_when_fileMissing_returnsDefaultsNoError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	cfg, err := LoadConfig(path)

	if err != nil {
		t.Fatalf("LoadConfig(%q) unexpected error: %v", path, err)
	}
	if want := DefaultConfig(); cfg != want {
		t.Errorf("LoadConfig(missing) = %+v, want default %+v", cfg, want)
	}
}

// Test_LoadConfig_when_pathEmpty_returnsDefaultsNoError
// 空路径 → os.ReadFile 报 not-exist → 与文件缺失同路径（默认值、无错误）。
func Test_LoadConfig_when_pathEmpty_returnsDefaultsNoError(t *testing.T) {
	cfg, err := LoadConfig("")

	if err != nil {
		t.Fatalf("LoadConfig(\"\") unexpected error: %v", err)
	}
	if want := DefaultConfig(); cfg != want {
		t.Errorf("LoadConfig(\"\") = %+v, want default %+v", cfg, want)
	}
}

// Test_LoadConfig_when_jsonCorrupt_returnsError
// 损坏 JSON → 返回明确 error（不 panic），且不覆盖原文件。
func Test_LoadConfig_when_jsonCorrupt_returnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{ not json"), 0o644); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}

	cfg, err := LoadConfig(path)

	if err == nil {
		t.Fatalf("LoadConfig(corrupt) returned nil error, cfg=%+v", cfg)
	}
}

// Test_LoadConfig_when_partialJSON_mergesOntoDefaults
// 文件存在但字段缺失：以默认值为底，文件中出现的字段覆盖对应默认值。
func Test_LoadConfig_when_partialJSON_mergesOntoDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	partial := `{"port":9090,"concurrency":4,"kvType":"fp8"}`
	if err := os.WriteFile(path, []byte(partial), 0o644); err != nil {
		t.Fatalf("write partial file: %v", err)
	}

	cfg, err := LoadConfig(path)

	if err != nil {
		t.Fatalf("LoadConfig(partial) unexpected error: %v", err)
	}
	if cfg.Port != 9090 || cfg.Concurrency != 4 || cfg.KvType != "fp8" {
		t.Errorf("overridden fields wrong: Port=%d Concurrency=%d KvType=%q",
			cfg.Port, cfg.Concurrency, cfg.KvType)
	}
	if cfg.ListenHost != "127.0.0.1" || cfg.Context != 131072 || cfg.Budget != 36864 {
		t.Errorf("untouched fields not defaulted: ListenHost=%q Context=%d Budget=%d",
			cfg.ListenHost, cfg.Context, cfg.Budget)
	}
}

// Test_LoadConfig_when_pathIsDirectory_returnsError
// 非 not-exist 的读取错误（目录）→ 返回 error。
func Test_LoadConfig_when_pathIsDirectory_returnsError(t *testing.T) {
	dir := t.TempDir()

	if _, err := LoadConfig(dir); err == nil {
		t.Fatalf("LoadConfig(dir) expected error, got nil")
	}
}
