// argv_llamacpp.go llama.cpp 引擎的 argv 构造契约。
//
// 与 ninfer 的 BuildArgs（argv_build.go）完全分离：flag 集、顺序、
// 默认值都不同。逐字对齐 start-backend.ps1 的 llamacpp 分支：
//
//  1. 固定段：-m Model --host H --port N -c Context -n MaxTokens
//     --kvmem-budget B --kvmem-gen-reserve R --kv-dtype T
//     -b Prefill -ub Prefill --kvmem-sink-tokens S --kvmem-recent-tokens Rc
//  2. 条件段：HostMiB>0 或 DiskPath 非空→--kvmem-host-mib；
//     DiskPath 非空→--kvmem-disk-path --kvmem-disk-mib；
//     !NoLrfu→--kvmem-lrfu --kvmem-lrfu-lambda --kvmem-lrfu-top-frac
//  3. 恒出现尾段：--kvmem-score-agg MODE
//
// 纯函数：无副作用，不涉及 os/exec。
package argv

import (
	"strconv"

	"ninferrmanager/internal/launcher/config"
)

// BuildLlamaArgs 构造 llama.cpp worker 的命令行参数（不含可执行文件路径）。
// 顺序固定（单测锁定），float 取值用 formatFloat（0.5→"0.5"）。
func BuildLlamaArgs(cfg config.LlamaConfig) []string {
	args := make([]string, 0, 32)

	// 1. 固定段：顺序与 flag 名逐字锁定
	args = append(args,
		"-m", cfg.ModelPath,
		"--host", cfg.Host,
		"--port", strconv.Itoa(cfg.Port),
		"-c", strconv.Itoa(cfg.Context),
		"-n", strconv.Itoa(cfg.MaxTokens),
		"--kvmem-budget", strconv.Itoa(cfg.Budget),
		"--kvmem-gen-reserve", strconv.Itoa(cfg.Reserve),
		"--kv-dtype", cfg.KvType,
		"-b", strconv.Itoa(cfg.Prefill),
		"-ub", strconv.Itoa(cfg.Prefill),
		"--kvmem-sink-tokens", strconv.Itoa(cfg.SinkTokens),
		"--kvmem-recent-tokens", strconv.Itoa(cfg.RecentTokens),
	)

	// 2. 条件段：出现顺序逐字锁定
	if cfg.HostMiB > 0 || cfg.DiskPath != "" {
		args = append(args, "--kvmem-host-mib", strconv.Itoa(cfg.HostMiB))
	}
	if cfg.DiskPath != "" {
		args = append(args, "--kvmem-disk-path", cfg.DiskPath, "--kvmem-disk-mib", strconv.Itoa(cfg.DiskMiB))
	}
	if !cfg.NoLrfu {
		args = append(args,
			"--kvmem-lrfu",
			"--kvmem-lrfu-lambda", formatFloat(cfg.LrfuLambda),
			"--kvmem-lrfu-top-frac", formatFloat(cfg.LrfuTopFraction),
		)
	}

	// 3. 恒出现尾段
	args = append(args, "--kvmem-score-agg", cfg.ScoreAggregation)

	return args
}
