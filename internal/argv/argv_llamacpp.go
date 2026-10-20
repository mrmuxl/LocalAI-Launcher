// argv_llamacpp.go llama.cpp 引擎的 argv 构造契约。
//
// 与 ninfer 的 BuildArgs（argv_ninfer_build.go）完全分离：flag 集、顺序、
// 默认值都不同。逐字对齐 H:\models\Qwen3.8-27B-GSQ_S-MTP.cmd 启动脚本：
//
//  1. 固定段：--kv-unified -m Model --host --port -t -c -n --gpu-layers
//     -b -ub --gpu-layers-draft -fa --parallel --keep --load-mode --reasoning
//     --cache-type-v/k(-draft) --spec-type --spec-draft-n-min/max
//     --ctx-checkpoints --poll --cache-ram --temp --top-p --top-k --min-p
//     --presence-penalty --repeat-penalty
//  2. 布尔开关段：--mmproj-offload --jinja --reasoning-preserve
//     --no-cache-idle-slots --no-host --no-warmup
//  3. 条件段：MmprojPath 非空→--mmproj；ChatTemplateFile 非空→
//     --chat-template-file；ChatTemplateKwargs 非空→--chat-template-kwargs
//  4. 数值条件段：ImageMinTokens/ImageMaxTokens 恒传（脚本默认 1024/4096）
//
// 纯函数：无副作用，不涉及 os/exec。
package argv

import (
	"strconv"

	"llmlauncher/internal/config"
)

// BuildLlamaArgs 构造 llama.cpp worker 的命令行参数（不含可执行文件路径）。
// 顺序固定（单测锁定），float 取值用 formatFloat（0.95→"0.95"）。
func BuildLlamaArgs(cfg config.LlamaConfig) []string {
	args := make([]string, 0, 64)

	// 1. 固定段：顺序与 flag 名逐字锁定（对齐 cmd 脚本续行顺序）
	if cfg.StartMode == "route" {
		// route 模式：模型由预设文件（--models-preset）决定，不传 -m。
		args = append(args, "--models-preset", cfg.ModelsPreset)
	} else {
		args = append(args, "-m", cfg.ModelPath)
	}
	args = append(args,
		"--alias", cfg.Alias,
		"--host", cfg.Host,
		"--port", strconv.Itoa(cfg.Port),
		"-t", strconv.Itoa(cfg.Threads),
		"-c", strconv.Itoa(cfg.CtxSize),
		"-n", strconv.Itoa(cfg.NPredict),
		"--gpu-layers", strconv.Itoa(cfg.GpuLayers),
		"-b", strconv.Itoa(cfg.BatchSize),
		"-ub", strconv.Itoa(cfg.UbatchSize),
		"--gpu-layers-draft", cfg.GpuLayersDraft,
		"-fa", cfg.FlashAttn,
		"--parallel", strconv.Itoa(cfg.Parallel),
		"--reasoning-format", cfg.ReasoningFormat,
		"-fit", cfg.Fit,
		"--keep", strconv.Itoa(cfg.Keep),
		"--load-mode", cfg.LoadMode,
		"--reasoning", cfg.Reasoning,
		"--cache-type-v", cfg.CacheTypeV,
		"--cache-type-k", cfg.CacheTypeK,
		"--cache-type-v-draft", cfg.CacheTypeVDraft,
		"--cache-type-k-draft", cfg.CacheTypeKDraft,
		"--spec-type", cfg.SpecType,
		"--spec-draft-n-min", strconv.Itoa(cfg.SpecDraftNMin),
		"--spec-draft-n-max", strconv.Itoa(cfg.SpecDraftNMax),
		"--ctx-checkpoints", strconv.Itoa(cfg.CtxCheckpoints),
		"--poll", strconv.Itoa(cfg.Poll),
		"--cache-ram", strconv.Itoa(cfg.CacheRam),
		"--temp", formatFloat(cfg.Temp),
		"--top-p", formatFloat(cfg.TopP),
		"--top-k", strconv.Itoa(cfg.TopK),
		"--min-p", formatFloat(cfg.MinP),
		"--presence-penalty", formatFloat(cfg.PresencePenalty),
		"--repeat-penalty", formatFloat(cfg.RepeatPenalty),
	)

	// 2. 布尔开关段：true 时追加 flag（无值）
	if cfg.KvUnified {
		args = append(args, "--kv-unified")
	}
	if cfg.MmprojOffload {
		args = append(args, "--mmproj-offload")
	}
	if cfg.Jinja {
		args = append(args, "--jinja")
	}
	if cfg.ReasoningPreserve {
		args = append(args, "--reasoning-preserve")
	}
	if cfg.NoCacheIdleSlots {
		args = append(args, "--no-cache-idle-slots")
	}
	if cfg.NoHost {
		args = append(args, "--no-host")
	}
	if cfg.NoWarmup {
		args = append(args, "--no-warmup")
	}

	// 3. 条件段：路径/模板非空才传
	// route 模式：mmproj 由预设文件决定，不传 --mmproj。
	if cfg.MmprojPath != "" && cfg.StartMode != "route" {
		args = append(args, "--mmproj", cfg.MmprojPath)
	}
	if cfg.ImageMinTokens > 0 {
		args = append(args, "--image-min-tokens", strconv.Itoa(cfg.ImageMinTokens))
	}
	if cfg.ImageMaxTokens > 0 {
		args = append(args, "--image-max-tokens", strconv.Itoa(cfg.ImageMaxTokens))
	}
	if cfg.ChatTemplateFile != "" {
		args = append(args, "--chat-template-file", cfg.ChatTemplateFile)
	}
	if cfg.ChatTemplateKwargs != "" {
		args = append(args, "--chat-template-kwargs", cfg.ChatTemplateKwargs)
	}

	return args
}
