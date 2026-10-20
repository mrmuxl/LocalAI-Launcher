package argv

import (
	"slices"
	"testing"

	"ninferrmanager/internal/launcher/config"
)

// TestBuildLlamaArgs_fixedSegmentOrder 固定段 flag 名与顺序逐字锁定
// （对齐 start-backend.ps1 llamacpp 分支）。
func TestBuildLlamaArgs_fixedSegmentOrder(t *testing.T) {
	cfg := config.DefaultLlamaConfig()
	cfg.ModelPath = `D:\models\qwen.gguf`

	got := BuildLlamaArgs(cfg)
	want := []string{
		"-m", `D:\models\qwen.gguf`,
		"--host", "127.0.0.1",
		"--port", "18200",
		"-c", "262144",
		"-n", "1024",
		"--kvmem-budget", "32768",
		"--kvmem-gen-reserve", "1024",
		"--kv-dtype", "q8_0",
		"-b", "128",
		"-ub", "128",
		"--kvmem-sink-tokens", "12288",
		"--kvmem-recent-tokens", "1024",
		"--kvmem-host-mib", "16384",
		"--kvmem-lrfu",
		"--kvmem-lrfu-lambda", "0.5",
		"--kvmem-lrfu-top-frac", "0.9",
		"--kvmem-score-agg", "max",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("argv mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

// TestBuildLlamaArgs_conditionals 条件段：DiskPath 非空带 disk 段；
// NoLrfu 关闭 LRFU 段；HostMiB=0 且无 DiskPath 时不传 host-mib。
func TestBuildLlamaArgs_conditionals(t *testing.T) {
	cfg := config.DefaultLlamaConfig()
	cfg.DiskPath = `D:\kv-cache`
	cfg.DiskMiB = 8192
	got := BuildLlamaArgs(cfg)
	for _, want := range []string{"--kvmem-disk-path", `D:\kv-cache`, "--kvmem-disk-mib", "8192"} {
		if !slices.Contains(got, want) {
			t.Errorf("disk 段缺少 %q", want)
		}
	}

	cfg2 := config.DefaultLlamaConfig()
	cfg2.NoLrfu = true
	cfg2.HostMiB = 0
	got2 := BuildLlamaArgs(cfg2)
	if slices.Contains(got2, "--kvmem-lrfu") {
		t.Error("NoLrfu 时不应出现 --kvmem-lrfu")
	}
	if slices.Contains(got2, "--kvmem-host-mib") {
		t.Error("HostMiB=0 且无 DiskPath 时不应出现 --kvmem-host-mib")
	}
}
