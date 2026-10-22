package argv

import (
	"slices"
	"testing"

	"llmlauncher/internal/config"
)

// TestBuildLlamaArgs_fixedSegmentOrder 固定段 flag 名与顺序逐字锁定
// （对齐 Qwen3.8-27B-GSQ_S-MTP.cmd 启动脚本续行顺序）。
func TestBuildLlamaArgs_fixedSegmentOrder(t *testing.T) {
	cfg := config.DefaultLlamaConfig()
	cfg.ModelPath = `D:\models\qwen.gguf`

	got := BuildLlamaArgs(cfg)
	want := []string{
		"-m", `D:\models\qwen.gguf`,
		"--alias", "Qwen3.8-27B",
		"--host", "0.0.0.0",
		"--port", "8080",
		"-t", "4",
		"-c", "81920",
		"-n", "16384",
		"--gpu-layers", "99",
		"-b", "512",
		"-ub", "512",
		"--gpu-layers-draft", "all",
		"-fa", "on",
		"--parallel", "1",
		"--reasoning-format", "deepseek",
		"-fit", "off",
		"--keep", "-1",
		"--load-mode", "mlock",
		"--reasoning", "on",
		"--cache-type-v", "q4_0",
		"--cache-type-k", "q4_0",
		"--cache-type-v-draft", "q4_0",
		"--cache-type-k-draft", "q4_0",
		"--spec-type", "draft-mtp",
		"--spec-draft-n-min", "1",
		"--spec-draft-n-max", "3",
		"--ctx-checkpoints", "4",
		"--poll", "10",
		"--cache-ram", "8192",
		"--temp", "1",
		"--top-p", "0.95",
		"--top-k", "20",
		"--min-p", "0",
		"--presence-penalty", "0",
		"--repeat-penalty", "1",
		"--kv-unified",
		"--mmproj-offload",
		"--jinja",
		"--reasoning-preserve",
		"--no-cache-idle-slots",
		"--no-host",
		"--no-warmup",
		"--mmproj", `H:\models\Qwen3.8-GSQ\mmproj-Qwen3.8-27B-BF16.gguf`,
		"--image-min-tokens", "1024",
		"--image-max-tokens", "4096",
		"--chat-template-file", `Qwen-Sharp-Chat-Templates\chat_template.jinja`,
		"--chat-template-kwargs", `{"preserve-thinking": true,"reasoning_effort": "medium"}`,
	}
	if !slices.Equal(got, want) {
		t.Fatalf("argv mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

// TestBuildLlamaArgs_conditionals 条件段：MmprojPath/ChatTemplateFile/
// ChatTemplateKwargs 为空时不传对应 flag；布尔开关关闭时不出现。
func TestBuildLlamaArgs_conditionals(t *testing.T) {
	cfg := config.DefaultLlamaConfig()
	cfg.MmprojPath = ""
	cfg.ChatTemplateFile = ""
	cfg.ChatTemplateKwargs = ""
	cfg.KvUnified = false
	cfg.MmprojOffload = false
	cfg.Jinja = false
	cfg.ReasoningPreserve = false
	cfg.NoCacheIdleSlots = false
	cfg.NoHost = false
	cfg.NoWarmup = false
	cfg.ImageMinTokens = 0
	cfg.ImageMaxTokens = 0

	got := BuildLlamaArgs(cfg)
	for _, banned := range []string{
		"--mmproj", "--image-min-tokens", "--image-max-tokens",
		"--chat-template-file", "--chat-template-kwargs",
		"--kv-unified", "--mmproj-offload", "--jinja", "--reasoning-preserve",
		"--no-cache-idle-slots", "--no-host", "--no-warmup",
	} {
		if slices.Contains(got, banned) {
			t.Errorf("条件关闭时不应出现 %q", banned)
		}
	}
}

// denseBaseline 默认配置（dense、无 MoE flag）的完整 argv 快照，
// 与 TestBuildLlamaArgs_fixedSegmentOrder 的 want 一致；MoE 用例在其上
// 拼接 MoE flag 做全量快照断言。
var denseBaseline = []string{
	"-m", `D:\models\qwen.gguf`,
	"--alias", "Qwen3.8-27B",
	"--host", "0.0.0.0",
	"--port", "8080",
	"-t", "4",
	"-c", "81920",
	"-n", "16384",
	"--gpu-layers", "99",
	"-b", "512",
	"-ub", "512",
	"--gpu-layers-draft", "all",
	"-fa", "on",
	"--parallel", "1",
	"--reasoning-format", "deepseek",
	"-fit", "off",
	"--keep", "-1",
	"--load-mode", "mlock",
	"--reasoning", "on",
	"--cache-type-v", "q4_0",
	"--cache-type-k", "q4_0",
	"--cache-type-v-draft", "q4_0",
	"--cache-type-k-draft", "q4_0",
	"--spec-type", "draft-mtp",
	"--spec-draft-n-min", "1",
	"--spec-draft-n-max", "3",
	"--ctx-checkpoints", "4",
	"--poll", "10",
	"--cache-ram", "8192",
	"--temp", "1",
	"--top-p", "0.95",
	"--top-k", "20",
	"--min-p", "0",
	"--presence-penalty", "0",
	"--repeat-penalty", "1",
	"--kv-unified",
	"--mmproj-offload",
	"--jinja",
	"--reasoning-preserve",
	"--no-cache-idle-slots",
	"--no-host",
	"--no-warmup",
	"--mmproj", `H:\models\Qwen3.8-GSQ\mmproj-Qwen3.8-27B-BF16.gguf`,
	"--image-min-tokens", "1024",
	"--image-max-tokens", "4096",
	"--chat-template-file", `Qwen-Sharp-Chat-Templates\chat_template.jinja`,
	"--chat-template-kwargs", `{"preserve-thinking": true,"reasoning_effort": "medium"}`,
}

// moeAfterFixedSegment 将 MoE flag 序列拼到固定段之后、布尔段
// （--kv-unified 等）之前，供全量快照断言。
func moeAfterFixedSegment(flags ...string) []string {
	idx := slices.Index(denseBaseline, "--kv-unified")
	out := make([]string, 0, len(denseBaseline)+len(flags))
	out = append(out, denseBaseline[:idx]...)
	out = append(out, flags...)
	return append(out, denseBaseline[idx:]...)
}

// moeBaseConfig 构建 MoE 用例的基础配置：默认值 + moe 模型类型。
func moeBaseConfig() config.LlamaConfig {
	cfg := config.DefaultLlamaConfig()
	cfg.ModelPath = `D:\models\qwen.gguf`
	cfg.MoeModelType = "moe"
	return cfg
}

// TestBuildLlamaArgs_moeCpuModeAll moe + all → 传 --cpu-moe；
// --n-cpu-moe/--moe-cache-mib 未设时不出现。
func TestBuildLlamaArgs_moeCpuModeAll(t *testing.T) {
	cfg := moeBaseConfig()
	cfg.MoeCpuMode = "all"

	got := BuildLlamaArgs(cfg)
	if !slices.Contains(got, "--cpu-moe") {
		t.Fatalf("应含 --cpu-moe, got %q", got)
	}
	for _, banned := range []string{"--n-cpu-moe", "--moe-cache-mib"} {
		if slices.Contains(got, banned) {
			t.Errorf("未设时不应出现 %q, got %q", banned, got)
		}
	}
}

// TestBuildLlamaArgs_moeFirstN_positive moe + firstN + N=5 → 传 --n-cpu-moe 5，
// 不传 --cpu-moe。
func TestBuildLlamaArgs_moeFirstN_positive(t *testing.T) {
	cfg := moeBaseConfig()
	cfg.MoeCpuMode = "firstN"
	cfg.NCpuMoe = 5

	got := BuildLlamaArgs(cfg)
	i := slices.Index(got, "--n-cpu-moe")
	if i < 0 || got[i+1] != "5" {
		t.Fatalf("应含 --n-cpu-moe 5 序列, got %q", got)
	}
	if slices.Contains(got, "--cpu-moe") {
		t.Errorf("firstN 模式不应传 --cpu-moe, got %q", got)
	}
}

// TestBuildLlamaArgs_moeFirstN_zero moe + firstN + N=0 → 不传 --n-cpu-moe
// （与实现中 NCpuMoe > 0 条件一致）。
func TestBuildLlamaArgs_moeFirstN_zero(t *testing.T) {
	cfg := moeBaseConfig()
	cfg.MoeCpuMode = "firstN"
	cfg.NCpuMoe = 0

	if got := BuildLlamaArgs(cfg); slices.Contains(got, "--n-cpu-moe") {
		t.Errorf("N=0 时不应传 --n-cpu-moe, got %q", got)
	}
}

// TestBuildLlamaArgs_moeCacheMb moe + MoeCacheMB=6144 → 传 --moe-cache-mib 6144。
func TestBuildLlamaArgs_moeCacheMb(t *testing.T) {
	cfg := moeBaseConfig()
	cfg.MoeCacheMB = 6144

	got := BuildLlamaArgs(cfg)
	i := slices.Index(got, "--moe-cache-mib")
	if i < 0 || got[i+1] != "6144" {
		t.Fatalf("应含 --moe-cache-mib 6144 序列, got %q", got)
	}
}

// TestBuildLlamaArgs_moeCombined_firstNCacheSnapshot 全量快照：MoE 三项设置
// （firstN + N + 缓存）全设，--n-cpu-moe/--moe-cache-mib 位于固定段之后、
// 布尔段（--kv-unified 等）之前。
func TestBuildLlamaArgs_moeCombined_firstNCacheSnapshot(t *testing.T) {
	cfg := moeBaseConfig()
	cfg.MoeCpuMode = "firstN"
	cfg.NCpuMoe = 5
	cfg.MoeCacheMB = 6144

	want := moeAfterFixedSegment("--n-cpu-moe", "5", "--moe-cache-mib", "6144")
	if got := BuildLlamaArgs(cfg); !slices.Equal(got, want) {
		t.Fatalf("argv mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

// TestBuildLlamaArgs_moeCombined_allCacheSnapshot 全量快照：all + 缓存，
// --cpu-moe/--moe-cache-mib 同样位于固定段之后、布尔段之前。
func TestBuildLlamaArgs_moeCombined_allCacheSnapshot(t *testing.T) {
	cfg := moeBaseConfig()
	cfg.MoeCpuMode = "all"
	cfg.MoeCacheMB = 6144

	want := moeAfterFixedSegment("--cpu-moe", "--moe-cache-mib", "6144")
	if got := BuildLlamaArgs(cfg); !slices.Equal(got, want) {
		t.Fatalf("argv mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

// TestBuildLlamaArgs_denseAndEmpty_noMoEFlags dense 与空 MoeModelType
// → 与改动前基线完全一致，无任何 MoE flag。
func TestBuildLlamaArgs_denseAndEmpty_noMoEFlags(t *testing.T) {
	for _, mode := range []string{"dense", ""} {
		cfg := config.DefaultLlamaConfig()
		cfg.ModelPath = `D:\models\qwen.gguf`
		cfg.MoeModelType = mode

		got := BuildLlamaArgs(cfg)
		for _, banned := range []string{"--cpu-moe", "--n-cpu-moe", "--moe-cache-mib"} {
			if slices.Contains(got, banned) {
				t.Errorf("mode %q 不应含 %q, got %q", mode, banned, got)
			}
		}
		if !slices.Equal(got, denseBaseline) {
			t.Errorf("mode %q argv 应与基线逐字一致:\ngot:  %q\nwant: %q", mode, got, denseBaseline)
		}
	}
}
