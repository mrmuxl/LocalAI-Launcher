package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Test_SaveConfig_when_roundTrip_preservesAll32Fields
// 写入临时目录 → LoadConfig 读回 → 整结构相等（SaveConfig/LoadConfig JSON 往返）。
func Test_SaveConfig_when_roundTrip_preservesAll32Fields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	cfg := DefaultConfig()
	cfg.ListenHost = "0.0.0.0"
	cfg.Port = 9090
	cfg.KvType = "fp8"
	cfg.FastPrefill = true
	cfg.Vision = false

	if err := SaveConfig(cfg, path); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	got, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig after SaveConfig: %v", err)
	}
	if got != cfg {
		t.Errorf("round trip mismatch:\ngot:  %+v\nwant: %+v", got, cfg)
	}
}

// Test_SaveConfig_when_parentDirsMissing_createsThem
func Test_SaveConfig_when_parentDirsMissing_createsThem(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "c", "config.json")

	if err := SaveConfig(DefaultConfig(), path); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}
	if len(data) == 0 {
		t.Error("saved file is empty")
	}
}

// Test_SaveConfig_when_pathComponentIsFile_returnsError
// 父路径中存在同名普通文件 → MkdirAll 失败 → 明确错误，不 panic。
func Test_SaveConfig_when_pathComponentIsFile_returnsError(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("creating blocker file: %v", err)
	}

	err := SaveConfig(DefaultConfig(), filepath.Join(blocker, "config.json"))
	if err == nil {
		t.Fatal("SaveConfig under a regular file returned nil error, want error")
	}
}
