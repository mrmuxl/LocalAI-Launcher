package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Test_SaveJSONAtomic_when_written_contentMatchesAndNoTmpLeft
// 正常写：目标文件内容一致，且 .tmp 临时文件已被 rename 清理。
func Test_SaveJSONAtomic_when_written_contentMatchesAndNoTmpLeft(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.json")
	data := []byte(`{"workerPath":"x"}`)

	if err := saveJSONAtomic(path, data); err != nil {
		t.Fatalf("saveJSONAtomic: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("回读目标文件: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("内容不一致: got %s, want %s", got, data)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("写成功后不应残留 .tmp 文件")
	}
}

// Test_SaveJSONAtomic_when_targetDirMissing_returnsError
// 目标目录不存在 → 临时文件写入失败 → 返回错误而非 panic。
func Test_SaveJSONAtomic_when_targetDirMissing_returnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "cfg.json")

	if err := saveJSONAtomic(path, []byte("{}")); err == nil {
		t.Fatal("saveJSONAtomic 目标目录不存在时应返回错误, got nil")
	}
}
