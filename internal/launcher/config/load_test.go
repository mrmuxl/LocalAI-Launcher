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
	partial := `{"port":9090,"workers":8,"modelName":"custom-model"}`
	if err := os.WriteFile(path, []byte(partial), 0o644); err != nil {
		t.Fatalf("write partial file: %v", err)
	}

	cfg, err := LoadConfig(path)

	if err != nil {
		t.Fatalf("LoadConfig(partial) unexpected error: %v", err)
	}
	if cfg.Port != 9090 || cfg.Workers != 8 || cfg.ModelName != "custom-model" {
		t.Errorf("overridden fields wrong: Port=%d Workers=%d ModelName=%q",
			cfg.Port, cfg.Workers, cfg.ModelName)
	}
	if cfg.Host != "127.0.0.1" || cfg.LogLevel != "info" || cfg.CtxSize != 32768 {
		t.Errorf("untouched fields not defaulted: Host=%q LogLevel=%q CtxSize=%d",
			cfg.Host, cfg.LogLevel, cfg.CtxSize)
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

// Test_ConfigPathFor_when_workerPathGiven_returnsSameDir
// 配置文件跟随推理程序：绝对路径 → 同目录指定文件名；相对路径 → 工作目录。
func Test_ConfigPathFor_when_workerPathGiven_returnsSameDir(t *testing.T) {
	abs := filepath.Join("D:", "ninfer", "ninfer.exe")
	if got := ConfigPathFor(abs, "config.json"); got != filepath.Join("D:", "ninfer", "config.json") {
		t.Errorf("ConfigPathFor(abs) = %q, want same dir config.json", got)
	}
	if got := ConfigPathFor(abs, "llama.json"); got != filepath.Join("D:", "ninfer", "llama.json") {
		t.Errorf("ConfigPathFor(abs, llama.json) = %q, want same dir llama.json", got)
	}
	if got := ConfigPathFor("ninfer.exe", "config.json"); got != "config.json" {
		t.Errorf("ConfigPathFor(relative) = %q, want config.json in work dir", got)
	}
}
