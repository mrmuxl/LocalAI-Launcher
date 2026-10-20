package argv

import (
	"strconv"
	"strings"

	"llmlauncher/internal/config"
)

// BuildArgs 按真实命令行契约构造 ninfer-kvmem-server.exe 参数（不含 exe 路径）。
//
// 逐字对齐 start-backend.ps1 的 ninfer 分支（顺序锁定，单测锁定）：
//
//  1. 位置参数：ModelPath
//  2. 固定段：--host --port --max-context --max-concurrency --prefill-chunk
//     --default-max-tokens --kvmem-budget --kvmem-gen-reserve --kvmem-host-mib
//     --kvmem-sink-tokens --kvmem-recent-tokens --kvmem-sessions --kv-dtype
//     --device-profile
//  3. 设备段：单卡 --device 0；StageLayers 非空 → --devices 0..n-1 + --stage-layers
//  4. 推测解码段：MtpDrafts>0 → --spec mtp --draft-tokens --ngram-draft-tokens
//     --ngram-min-match；AdaptiveMtp → --adaptive-mtp
//  5. FastPrefill → --fast-prefill-kernel
//  6. Vision → --vision --vision-residency --vision-max-merged
//  7. ThinkingBudget>0 → --default-thinking-budget
//  8. DiskPath 非空 → --kvmem-disk-path --kvmem-disk-mib
//  9. LRFU 段：!NoLrfu → --kvmem-lrfu --kvmem-lrfu-lambda --kvmem-lrfu-top-frac
//  10. 尾段：--kvmem-score-agg（恒现）
//
// Gpu 不进 argv：经环境变量 CUDA_VISIBLE_DEVICES 传递（process.Manager.SetEnv）。
// 纯函数：无副作用，不涉及 os/exec。
// float 取值一律 formatFloat（strconv.FormatFloat(v,'g',-1,64)），0.5→"0.5"。
func BuildArgs(cfg config.Config) []string {
	args := make([]string, 0, 48)

	// 1. 位置参数：模型文件
	args = append(args, cfg.ModelPath)

	// 2. 固定段：顺序与 flag 名逐字锁定
	args = append(args,
		"--host", cfg.ListenHost,
		"--port", strconv.Itoa(cfg.Port),
		"--max-context", strconv.Itoa(cfg.Context),
		"--max-concurrency", strconv.Itoa(cfg.Concurrency),
		"--prefill-chunk", strconv.Itoa(cfg.Prefill),
		"--default-max-tokens", strconv.Itoa(cfg.MaxTokens),
		"--kvmem-budget", strconv.Itoa(cfg.Budget),
		"--kvmem-gen-reserve", strconv.Itoa(cfg.Reserve),
		"--kvmem-host-mib", strconv.Itoa(cfg.HostMiB),
		"--kvmem-sink-tokens", strconv.Itoa(cfg.SinkTokens),
		"--kvmem-recent-tokens", strconv.Itoa(cfg.RecentTokens),
		"--kvmem-sessions", strconv.Itoa(cfg.RetainedSessions),
		"--kv-dtype", cfg.KvType,
		"--device-profile", cfg.DeviceProfile,
	)

	// 3. 设备段：单卡 --device 0；多卡（StageLayers 非空）--devices 0..n-1
	if cfg.StageLayers != "" {
		n := len(strings.Split(cfg.StageLayers, ","))
		idx := make([]string, n)
		for i := range idx {
			idx[i] = strconv.Itoa(i)
		}
		args = append(args, "--devices", strings.Join(idx, ","))
		args = append(args, "--stage-layers", cfg.StageLayers)
	} else {
		args = append(args, "--device", "0")
	}

	// 4. 推测解码段
	if cfg.MtpDrafts > 0 {
		args = append(args,
			"--spec", "mtp",
			"--draft-tokens", strconv.Itoa(cfg.MtpDrafts),
			"--ngram-draft-tokens", strconv.Itoa(cfg.NgramDrafts),
			"--ngram-min-match", strconv.Itoa(cfg.NgramMinMatch),
		)
	}
	if cfg.AdaptiveMtp {
		args = append(args, "--adaptive-mtp")
	}

	// 5. 快速 prefill 内核
	if cfg.FastPrefill {
		args = append(args, "--fast-prefill-kernel")
	}

	// 6. 视觉段
	if cfg.Vision {
		args = append(args,
			"--vision",
			"--vision-residency", cfg.VisionResidency,
			"--vision-max-merged", strconv.Itoa(cfg.VisionTokens),
		)
	}

	// 7. 思考预算与思考力度
	if cfg.ThinkingBudget > 0 {
		args = append(args, "--default-thinking-budget", strconv.Itoa(cfg.ThinkingBudget))
	}
	if cfg.DefaultReasoningEffort != "" {
		args = append(args, "--default-reasoning-effort", cfg.DefaultReasoningEffort)
	}

	// 8. SSD 分层段
	if cfg.DiskPath != "" {
		args = append(args, "--kvmem-disk-path", cfg.DiskPath, "--kvmem-disk-mib", strconv.Itoa(cfg.DiskMiB))
	}

	// 9. LRFU 段（!NoLrfu 时出现）
	if !cfg.NoLrfu {
		args = append(args,
			"--kvmem-lrfu",
			"--kvmem-lrfu-lambda", formatFloat(cfg.LrfuLambda),
			"--kvmem-lrfu-top-frac", formatFloat(cfg.LrfuTopFraction),
		)
	}

	// 10. 尾段：评分聚合（恒现）
	args = append(args, "--kvmem-score-agg", cfg.ScoreAggregation)

	return args
}
