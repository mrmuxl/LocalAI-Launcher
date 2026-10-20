package gui

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/test"
)

// TestFilePicker_matchExt 扩展名过滤：空 exts 全部通过；含点小写匹配、大小写不敏感。
func TestFilePicker_matchExt(t *testing.T) {
	p := &filePicker{}

	p.exts = nil
	if !p.matchExt("a.gguf") || !p.matchExt("b.exe") {
		t.Fatal("空 exts 应全部通过")
	}

	p.exts = []string{".gguf"}
	if !p.matchExt("model.GGUF") {
		t.Fatal("大小写不敏感匹配失败")
	}
	if p.matchExt("model.ninfer") || p.matchExt("modelgguf") {
		t.Fatal("不应匹配非 .gguf 文件")
	}
}

// TestFilePicker_enter 目录读取：目录在前文件在后、各自按名称排序、扩展名过滤生效。
func TestFilePicker_enter(t *testing.T) {
	test.NewApp()
	dir := t.TempDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("准备目录失败: %v", err)
		}
	}
	must(os.Mkdir(filepath.Join(dir, "zsub"), 0o755))
	must(os.Mkdir(filepath.Join(dir, "asub"), 0o755))
	must(os.WriteFile(filepath.Join(dir, "b.gguf"), []byte("x"), 0o644))
	must(os.WriteFile(filepath.Join(dir, "a.gguf"), []byte("x"), 0o644))
	must(os.WriteFile(filepath.Join(dir, "skip.txt"), []byte("x"), 0o644))

	p := newFilePicker(test.NewApp(), "测试", dir, []string{".gguf"}, func(string, string) {})
	if p.cur != dir {
		t.Fatalf("cur 不一致: want %q, got %q", dir, p.cur)
	}
	want := []dirEntry{
		{name: "asub", isDir: true},
		{name: "zsub", isDir: true},
		{name: "a.gguf"},
		{name: "b.gguf"},
	}
	if len(p.entries) != len(want) {
		t.Fatalf("条目数不一致: want %d, got %d (%+v)", len(want), len(p.entries), p.entries)
	}
	for i, w := range want {
		if p.entries[i] != w {
			t.Fatalf("第 %d 项不一致: want %+v, got %+v", i, w, p.entries[i])
		}
	}
}

// TestFilePicker_confirm 确认逻辑：手输存在的文件名回调 onPick 并关窗；
// 不存在的文件名不回调；目录名不回调。
func TestFilePicker_confirm(t *testing.T) {
	test.NewApp()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "model.gguf"), []byte("x"), 0o644); err != nil {
		t.Fatalf("准备文件失败: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatalf("准备目录失败: %v", err)
	}

	var picked string
	p := newFilePicker(test.NewApp(), "测试", dir, []string{".gguf"}, func(path, _ string) { picked = path })

	p.nameBox.SetText("missing.gguf")
	p.confirm()
	if picked != "" {
		t.Fatalf("不存在文件不应回调, got %q", picked)
	}

	p.nameBox.SetText("sub")
	p.confirm()
	if picked != "" {
		t.Fatalf("目录不应回调, got %q", picked)
	}

	p.nameBox.SetText("model.gguf")
	p.confirm()
	want := filepath.Join(dir, "model.gguf")
	if picked != want {
		t.Fatalf("回调路径不一致: want %q, got %q", want, picked)
	}
}
