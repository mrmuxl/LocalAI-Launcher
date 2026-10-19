package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadWorkerPathMissing 存档文件不存在 → 空串且无错误（首次使用语义）。
func TestLoadWorkerPathMissing(t *testing.T) {
	p, err := loadWorkerPathFrom(filepath.Join(t.TempDir(), "worker.json"))
	if err != nil {
		t.Fatalf("期望无错误, got %v", err)
	}
	if p != "" {
		t.Fatalf("期望空路径, got %q", p)
	}
}

// TestSaveLoadWorkerPathRoundtrip 保存后读取应还原出相同路径。
func TestSaveLoadWorkerPathRoundtrip(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "worker.json")

	want := `C:\tools\ninfer.exe`
	if err := saveWorkerPathTo(file, want); err != nil {
		t.Fatalf("保存失败: %v", err)
	}

	got, err := loadWorkerPathFrom(file)
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	if got != want {
		t.Fatalf("roundtrip 不一致: want %q, got %q", want, got)
	}
}

// TestSaveLoadWorkerPathRoundtripNested 保存时父目录不存在应自动创建。
func TestSaveLoadWorkerPathRoundtripNested(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "nested", "worker.json")

	want := `/opt/ninfer/ninfer`
	if err := saveWorkerPathTo(file, want); err != nil {
		t.Fatalf("保存失败: %v", err)
	}

	got, err := loadWorkerPathFrom(file)
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	if got != want {
		t.Fatalf("roundtrip 不一致: want %q, got %q", want, got)
	}
}

// TestLoadWorkerPathCorrupt JSON 损坏应返回错误且语义可读。
func TestLoadWorkerPathCorrupt(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "worker.json")
	if err := os.WriteFile(file, []byte("{bad json"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := loadWorkerPathFrom(file)
	if err == nil {
		t.Fatal("期望解析失败返回错误, got nil")
	}
	if !strings.Contains(err.Error(), "解析") {
		t.Fatalf("错误信息应包含「解析」, got %v", err)
	}
}
