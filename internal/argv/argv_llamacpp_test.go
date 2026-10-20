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
