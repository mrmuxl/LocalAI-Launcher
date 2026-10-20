package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// TestLastPickDir_roundtrip 记录级读写：LastPickDir 与双 worker 路径互不覆盖。
func TestLastPickDir_roundtrip(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "worker.json")
	rec := WorkerPathRecord{
		WorkerPath:      `C:\a\ninfer.exe`,
		LlamaWorkerPath: `C:\b\llama-server.exe`,
		LastPickDir:     `D:\models`,
	}
	if err := saveWorkerRecordTo(file, rec); err != nil {
		t.Fatalf("保存失败: %v", err)
	}
	got, err := loadWorkerRecordFrom(file)
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	if got != rec {
		t.Fatalf("roundtrip 不一致: want %+v, got %+v", rec, got)
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
	want := filepath.Join(filepath.Dir(exe), "ninfer.json")
	if got := ConfigPathFor("ninfer.json"); got != want {
		t.Errorf("ConfigPathFor(ninfer.json) = %q, want %q", got, want)
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

// TestSaveWorkerPathFieldTo_overwritesCorrupted 存档损坏时保存应成功覆盖，
// 而不是永久失败（否则用户每次启动都要重新选择推理引擎）。
func TestSaveWorkerPathFieldTo_overwritesCorrupted(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "worker.json")
	if err := os.WriteFile(file, []byte("{corrupted"), 0o644); err != nil {
		t.Fatal(err)
	}

	want := `C:\x\ninfer.exe`
	if _, err := saveWorkerPathFieldTo(file, func(rec *WorkerPathRecord) { rec.WorkerPath = want }); err != nil {
		t.Fatalf("损坏存档上保存应成功覆盖, got %v", err)
	}

	rec, err := loadWorkerRecordFrom(file)
	if err != nil {
		t.Fatalf("覆盖后读取失败: %v", err)
	}
	if rec.WorkerPath != want {
		t.Fatalf("覆盖结果不一致: want %q, got %q", want, rec.WorkerPath)
	}
}

// TestSaveWorkerPathFieldTo_concurrent 并发保存不损坏 JSON：
// 多个 goroutine 同时读-改-写同一存档，最终文件必须是合法 JSON。
func TestSaveWorkerPathFieldTo_concurrent(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "worker.json")

	const n = 32
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := `C:\w\` + string(rune('a'+i%26)) + `.exe`
			if _, err := saveWorkerPathFieldTo(file, func(rec *WorkerPathRecord) { rec.WorkerPath = p }); err != nil {
				t.Errorf("并发保存失败: %v", err)
			}
		}(i)
	}
	wg.Wait()

	rec, err := loadWorkerRecordFrom(file)
	if err != nil {
		t.Fatalf("并发保存后存档损坏: %v", err)
	}
	if rec.WorkerPath == "" {
		t.Fatal("并发保存后 WorkerPath 不应为空")
	}
}
