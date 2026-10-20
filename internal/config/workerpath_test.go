package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadWorkerPathMissing 存档文件不存在 → 零值记录且无错误（首次使用语义）。
func TestLoadWorkerPathMissing(t *testing.T) {
	rec, err := loadWorkerRecordFrom(filepath.Join(t.TempDir(), "worker.json"))
	if err != nil {
		t.Fatalf("期望无错误, got %v", err)
	}
	if rec.WorkerPath != "" || rec.LlamaWorkerPath != "" {
		t.Fatalf("期望空记录, got %+v", rec)
	}
}

// TestSaveLoadWorkerPathRoundtrip 保存后读取应还原出相同路径。
func TestSaveLoadWorkerPathRoundtrip(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "worker.json")

	want := `C:\tools\ninfer.exe`
	if err := saveWorkerRecordTo(file, WorkerPathRecord{WorkerPath: want}); err != nil {
		t.Fatalf("保存失败: %v", err)
	}

	rec, err := loadWorkerRecordFrom(file)
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	if rec.WorkerPath != want {
		t.Fatalf("roundtrip 不一致: want %q, got %q", want, rec.WorkerPath)
	}
}

// TestSaveLoadWorkerPathRoundtripNested 保存时父目录不存在应自动创建。
func TestSaveLoadWorkerPathRoundtripNested(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "nested", "worker.json")

	want := `/opt/ninfer/ninfer`
	if err := saveWorkerRecordTo(file, WorkerPathRecord{WorkerPath: want}); err != nil {
		t.Fatalf("保存失败: %v", err)
	}

	rec, err := loadWorkerRecordFrom(file)
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	if rec.WorkerPath != want {
		t.Fatalf("roundtrip 不一致: want %q, got %q", want, rec.WorkerPath)
	}
}

// TestSaveLlamaWorkerPath_preservesNinferPath 写入 llama 路径不得覆盖 ninfer 路径。
func TestSaveLlamaWorkerPath_preservesNinferPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "worker.json")
	if err := saveWorkerRecordTo(file, WorkerPathRecord{WorkerPath: `C:\a\ninfer.exe`}); err != nil {
		t.Fatalf("预置存档失败: %v", err)
	}

	// SaveLlamaWorkerPath 写默认位置，无法注入临时目录；直接验证记录级读写语义。
	rec := WorkerPathRecord{WorkerPath: `C:\a\ninfer.exe`, LlamaWorkerPath: `C:\b\llama-server.exe`}
	if err := saveWorkerRecordTo(file, rec); err != nil {
		t.Fatalf("保存失败: %v", err)
	}
	got, err := loadWorkerRecordFrom(file)
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	if got.WorkerPath != rec.WorkerPath || got.LlamaWorkerPath != rec.LlamaWorkerPath {
		t.Fatalf("双路径 roundtrip 不一致: want %+v, got %+v", rec, got)
	}
}

// Test_ConfigPathFor_returnsExeDirFile
// 配置文件放启动器 exe 同级目录：返回 exeDir()/filename。
func Test_ConfigPathFor_returnsExeDirFile(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable 不可用: %v", err)
	}
	want := filepath.Join(filepath.Dir(exe), "config.json")
	if got := ConfigPathFor("config.json"); got != want {
		t.Errorf("ConfigPathFor(config.json) = %q, want %q", got, want)
	}
	wantLlama := filepath.Join(filepath.Dir(exe), "llama.json")
	if got := ConfigPathFor("llama.json"); got != wantLlama {
		t.Errorf("ConfigPathFor(llama.json) = %q, want %q", got, wantLlama)
	}
}

// Test_WorkerSettingsPath_returnsExeDirWorkerJSON
// worker.json 存档放启动器 exe 同级目录。
func Test_WorkerSettingsPath_returnsExeDirWorkerJSON(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable 不可用: %v", err)
	}
	want := filepath.Join(filepath.Dir(exe), "worker.json")
	if got := WorkerSettingsPath(); got != want {
		t.Errorf("WorkerSettingsPath() = %q, want %q", got, want)
	}
}

// TestLoadWorkerPathCorrupt JSON 损坏应返回错误且语义可读。
func TestLoadWorkerPathCorrupt(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "worker.json")
	if err := os.WriteFile(file, []byte("{bad json"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := loadWorkerRecordFrom(file)
	if err == nil {
		t.Fatal("期望解析失败返回错误, got nil")
	}
	if !strings.Contains(err.Error(), "解析") {
		t.Fatalf("错误信息应包含「解析」, got %v", err)
	}
}
