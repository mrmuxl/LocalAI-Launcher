package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Validate 校验 ninfer 配置，返回中文错误文案列表（空=通过）。
// 规则逐条对齐 start-backend.ps1 的 ninfer 分支校验；含模型文件存在性
// 检查（fs IO），只可在 goroutine 中调用。
func Validate(cfg Config) []string {
	var errs []string

	// ── 基本范围（ps1 ValidateRange）────────
	if cfg.Port < 1 || cfg.Port > 65535 {
		errs = append(errs, fmt.Sprintf("监听端口 %d 超出范围 [1,65535]", cfg.Port))
	}
	if cfg.Concurrency < 1 || cfg.Concurrency > 8 {
		errs = append(errs, fmt.Sprintf("并发解码请求数 %d 超出范围 [1,8]", cfg.Concurrency))
	}
	if cfg.Context < 256 {
		errs = append(errs, fmt.Sprintf("上下文上限 %d 必须 ≥ 256", cfg.Context))
	}
	if cfg.Budget < 128 {
		errs = append(errs, fmt.Sprintf("历史预算 %d 必须 ≥ 128", cfg.Budget))
	}
	if cfg.Reserve < 64 {
		errs = append(errs, fmt.Sprintf("生成保留 %d 必须 ≥ 64", cfg.Reserve))
	}
	if cfg.HostMiB < 1 {
		errs = append(errs, fmt.Sprintf("Host KV 预算 %d 必须 ≥ 1", cfg.HostMiB))
	}
	if cfg.MaxTokens < 1 {
		errs = append(errs, fmt.Sprintf("默认输出上限 %d 必须 ≥ 1", cfg.MaxTokens))
	}
	if cfg.Prefill < 128 {
		errs = append(errs, fmt.Sprintf("prefill 块 %d 必须 ≥ 128", cfg.Prefill))
	}
	if cfg.RetainedSessions < 1 || cfg.RetainedSessions > 16 {
		errs = append(errs, fmt.Sprintf("保留历史数 %d 超出范围 [1,16]", cfg.RetainedSessions))
	}
	if cfg.MtpDrafts < 0 || cfg.MtpDrafts > 15 {
		errs = append(errs, fmt.Sprintf("MTP 草稿数 %d 超出范围 [0,15]", cfg.MtpDrafts))
	}
	if cfg.NgramDrafts < 0 || cfg.NgramDrafts > 63 {
		errs = append(errs, fmt.Sprintf("ngram 草稿数 %d 超出范围 [0,63]", cfg.NgramDrafts))
	}
	if cfg.NgramMinMatch < 4 || cfg.NgramMinMatch > 64 {
		errs = append(errs, fmt.Sprintf("ngram 最短匹配 %d 超出范围 [4,64]", cfg.NgramMinMatch))
	}
	if cfg.VisionTokens < 1 || cfg.VisionTokens > 16384 {
		errs = append(errs, fmt.Sprintf("视觉合并上限 %d 超出范围 [1,16384]", cfg.VisionTokens))
	}
	if cfg.ThinkingBudget < 0 {
		errs = append(errs, fmt.Sprintf("思考预算 %d 必须 ≥ 0", cfg.ThinkingBudget))
	}
	if cfg.DiskMiB < 1 {
		errs = append(errs, fmt.Sprintf("SSD 配额 %d 必须 ≥ 1", cfg.DiskMiB))
	}

	// ── 对齐与组合约束（ps1 ninfer 分支 throw）────────
	if cfg.Budget%64 != 0 {
		errs = append(errs, fmt.Sprintf("历史预算 %d 必须对齐 64", cfg.Budget))
	}
	if cfg.Reserve%64 != 0 {
		errs = append(errs, fmt.Sprintf("生成保留 %d 必须对齐 64", cfg.Reserve))
	}
	if cfg.Prefill%128 != 0 {
		errs = append(errs, fmt.Sprintf("prefill 块 %d 必须对齐 128", cfg.Prefill))
	}
	if cfg.Prefill > cfg.Reserve {
		errs = append(errs, fmt.Sprintf("prefill 块 %d 不得超过生成保留 %d", cfg.Prefill, cfg.Reserve))
	}
	if cfg.Budget+cfg.Reserve > cfg.Context {
		errs = append(errs, fmt.Sprintf("历史预算 + 生成保留（%d）不得超过上下文上限 %d", cfg.Budget+cfg.Reserve, cfg.Context))
	}
	if cfg.RecentTokens < 64 {
		errs = append(errs, fmt.Sprintf("常驻后缀 %d 必须 ≥ 64（保留最新一页）", cfg.RecentTokens))
	}
	sinkBlocks := int64(1)
	if cfg.SinkTokens >= 64 {
		sinkBlocks = int64(cfg.SinkTokens / 64)
	}
	recentBlocks := int64(cfg.RecentTokens / 64)
	if (sinkBlocks+recentBlocks)*64 > int64(cfg.Budget) {
		errs = append(errs, fmt.Sprintf("常驻前缀+后缀（%d token）按 64 对齐后超过历史预算 %d", (sinkBlocks+recentBlocks)*64, cfg.Budget))
	}

	// ── 枚举（ps1 ValidateSet / throw）────────
	kvTypes := map[string]bool{"bf16": true, "int8": true, "fp8": true, "nvfp4": true, "k8v4": true,
		"rk8v4": true, "rk4v4": true, "rk4v4-e8": true, "rk2v4-e8": true}
	if !kvTypes[cfg.KvType] {
		errs = append(errs, fmt.Sprintf("KV 存储 %q 无效（bf16/int8/fp8/nvfp4/k8v4/rk8v4/rk4v4/rk4v4-e8/rk2v4-e8）", cfg.KvType))
	}
	if cfg.VisionResidency != "resident" && cfg.VisionResidency != "cpu" {
		errs = append(errs, fmt.Sprintf("视觉驻留 %q 无效（resident/cpu）", cfg.VisionResidency))
	}
	if cfg.DeviceProfile != "off" && cfg.DeviceProfile != "auto" {
		errs = append(errs, fmt.Sprintf("设备档案 %q 无效（off/auto）", cfg.DeviceProfile))
	}
	if cfg.ScoreAggregation != "mean" && cfg.ScoreAggregation != "max" {
		errs = append(errs, fmt.Sprintf("评分聚合 %q 无效（mean/max）", cfg.ScoreAggregation))
	}

	// ── 组合约束（ps1 throw）────────
	if cfg.AdaptiveMtp && cfg.MtpDrafts == 0 {
		errs = append(errs, "自适应 MTP 需要先启用 MTP（草稿数 > 0）")
	}
	fastPrefillKv := map[string]bool{"int8": true, "rk8v4": true, "rk4v4": true, "rk4v4-e8": true, "rk2v4-e8": true}
	if cfg.FastPrefill && !fastPrefillKv[cfg.KvType] {
		errs = append(errs, fmt.Sprintf("快速 prefill 要求 KV 存储为 int8/rk8v4/rk4v4/rk4v4-e8/rk2v4-e8（当前 %s）", cfg.KvType))
	}
	if cfg.VisionResidency == "cpu" && !cfg.Vision {
		errs = append(errs, "CPU 视觉驻留需要先启用视觉")
	}
	if cfg.NgramDrafts > 0 {
		if cfg.MtpDrafts == 0 {
			errs = append(errs, "ngram 草稿需要先启用 MTP（草稿数 > 0）")
		} else {
			need := max(cfg.MtpDrafts, cfg.NgramDrafts) + cfg.MtpDrafts
			if need > cfg.Reserve {
				errs = append(errs, fmt.Sprintf("ngram 要求生成保留 ≥ %d（max(MTP,ngram)+MTP），当前 %d", need, cfg.Reserve))
			}
		}
	}

	// ── GPU 与多卡（ps1 ninfer 分支）────────
	if cfg.Gpu == "" {
		errs = append(errs, "GPU 不能为空（-Gpu / CUDA_VISIBLE_DEVICES）")
	} else {
		ids := strings.Split(cfg.Gpu, ",")
		seen := map[string]bool{}
		bad := len(ids) > 8
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id == "" || id == "-1" || seen[id] {
				bad = true
			}
			seen[id] = true
		}
		if bad {
			errs = append(errs, "GPU 需要 1..8 个互不重复的编号/UUID，逗号分隔，不能为 -1")
		}
	}
	if cfg.StageLayers != "" {
		parts := strings.Split(cfg.StageLayers, ",")
		if len(parts) < 2 {
			errs = append(errs, "每卡层数需要至少两张卡（逗号分隔）")
		}
		for _, p := range parts {
			v, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || v <= 0 {
				errs = append(errs, fmt.Sprintf("每卡层数 %q 无效（需要逗号分隔的正整数）", cfg.StageLayers))
				break
			}
		}
		if cfg.StageLayers != "" && len(strings.Split(cfg.Gpu, ",")) < 2 && len(strings.Split(cfg.StageLayers, ",")) >= 2 {
			errs = append(errs, "多卡层数需要 GPU 填写多个编号（逗号分隔）")
		}
	}

	// ── 模型文件存在性（fs IO）────────
	if cfg.ModelPath == "" {
		errs = append(errs, "模型文件不能为空")
	} else if st, err := os.Stat(cfg.ModelPath); err != nil {
		errs = append(errs, "模型文件不存在："+cfg.ModelPath)
	} else if st.IsDir() {
		errs = append(errs, "模型路径是目录，需要 .ninfer 文件："+cfg.ModelPath)
	}

	return errs
}
