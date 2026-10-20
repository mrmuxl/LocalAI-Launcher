package argv

import (
	"strconv"

	"llmlauncher/internal/config"
)

// BuildArgs 按计划「argv 构造契约」构造 worker 命令行参数（不含可执行文件路径）。
//
// 顺序固定（单测锁定），逐字实现：
//
//  1. 固定段（16 个 flag，恒出现）：
//     --bind Host --port Port --log-level LogLevel --workers Workers --model ModelName
//     --ctx CtxSize --threads Threads --quant Quant --temp Temperature --top-p TopP
//     --top-k TopK --repeat RepeatPenalty --max-tokens MaxTokens --cors CORSOrigins(逗号串)
//     --timeout TimeoutSec --max-conns MaxConns
//  2. 条件段（按契约顺序出现）：
//     APIKey 非空→--api-key；TLS→--tls --cert --key；
//     MetricsEnabled→--metrics [--metrics-path]（MetricsPath 非空才带路径）；
//     LogFile 非空→--log-file；MemLimitGB>0→--mem-limit-gb；SwapGB>0→--swap-gb；
//     NumaNode≥0→--numa；Priority≠normal→--priority；AffinityMask 非空→--affinity。
//  3. 恒出现缓存段：--cache-dir D --cache-mb N --evict P。
//  4. Prefetch→--prefetch（无值裸 flag）。
//
// 纯函数：无副作用，不涉及 os/exec。
// float 取值一律 formatFloat（strconv.FormatFloat(v,'g',-1,64)），0.7→"0.7"。
func BuildArgs(cfg config.Config) []string {
	args := make([]string, 0, 64)

	// 1. 固定段：顺序与 flag 名逐字锁定，一个不能多、不能少
	args = append(args,
		"--bind", cfg.Host,
		"--port", strconv.Itoa(cfg.Port),
		"--log-level", cfg.LogLevel,
		"--workers", strconv.Itoa(cfg.Workers),
		"--model", cfg.ModelName,
		"--ctx", strconv.Itoa(cfg.CtxSize),
		"--threads", strconv.Itoa(cfg.Threads),
		"--quant", cfg.Quant,
		"--temp", formatFloat(cfg.Temperature),
		"--top-p", formatFloat(cfg.TopP),
		"--top-k", strconv.Itoa(cfg.TopK),
		"--repeat", formatFloat(cfg.RepeatPenalty),
		"--max-tokens", strconv.Itoa(cfg.MaxTokens),
		"--cors", cfg.CORSOrigins,
		"--timeout", strconv.Itoa(cfg.TimeoutSec),
		"--max-conns", strconv.Itoa(cfg.MaxConns),
	)

	// 2. 条件段：出现顺序逐字锁定
	if cfg.APIKey != "" {
		args = append(args, "--api-key", cfg.APIKey)
	}
	if cfg.TLS {
		args = append(args, "--tls", "--cert", cfg.CertPath, "--key", cfg.KeyPath)
	}
	if cfg.MetricsEnabled {
		args = append(args, "--metrics")
		if cfg.MetricsPath != "" {
			args = append(args, "--metrics-path", cfg.MetricsPath)
		}
	}
	if cfg.LogFile != "" {
		args = append(args, "--log-file", cfg.LogFile)
	}
	if cfg.MemLimitGB > 0 {
		args = append(args, "--mem-limit-gb", strconv.Itoa(cfg.MemLimitGB))
	}
	if cfg.SwapGB > 0 {
		args = append(args, "--swap-gb", strconv.Itoa(cfg.SwapGB))
	}
	if cfg.NumaNode >= 0 {
		args = append(args, "--numa", strconv.Itoa(cfg.NumaNode))
	}
	if cfg.Priority != config.DefPriority {
		args = append(args, "--priority", cfg.Priority)
	}
	if cfg.AffinityMask != "" {
		args = append(args, "--affinity", cfg.AffinityMask)
	}

	// 3. 缓存段：恒出现
	args = append(args,
		"--cache-dir", cfg.CacheDir,
		"--cache-mb", strconv.Itoa(cfg.CacheSizeMB),
		"--evict", cfg.EvictPolicy,
	)

	// 4. Prefetch：无值裸 flag
	if cfg.Prefetch {
		args = append(args, "--prefetch")
	}

	return args
}
