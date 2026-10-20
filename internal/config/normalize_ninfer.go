package config

import "strings"

// Normalize 返回一份规整后的配置（不修改入参），规则 N1-N3，先于 Validate 执行。
//
//   - N1: 所有字符串字段 TrimSpace；CORSOrigins 按逗号拆分、去空格、去重（保序）后再拼接。
//   - N2: 可空字段（ModelPath/LogFile/CertPath/KeyPath/AffinityMask）保持空；Threads=0 保持 0（自动）。
//     本函数只规整不填默认值（默认值由 DefaultConfig/LoadConfig 负责）。
//   - N3: 数值下限——Port/Workers/CtxSize/Threads/MaxTokens/TimeoutSec/MaxConns ≥ 0（负数→0），
//     Temperature ≥ 0（负数→0）。TopP 不在此处钳制，越界交给 Validate 报错误。
func Normalize(cfg Config) Config {
	n := cfg

	// N1: TrimSpace 全部字符串字段
	n.Host = strings.TrimSpace(n.Host)
	n.LogLevel = strings.TrimSpace(n.LogLevel)
	n.ModelName = strings.TrimSpace(n.ModelName)
	n.ModelPath = strings.TrimSpace(n.ModelPath)
	n.Quant = strings.TrimSpace(n.Quant)
	n.APIKey = strings.TrimSpace(n.APIKey)
	n.CertPath = strings.TrimSpace(n.CertPath)
	n.KeyPath = strings.TrimSpace(n.KeyPath)
	n.MetricsPath = strings.TrimSpace(n.MetricsPath)
	n.LogFile = strings.TrimSpace(n.LogFile)
	n.Priority = strings.TrimSpace(n.Priority)
	n.AffinityMask = strings.TrimSpace(n.AffinityMask)
	n.CacheDir = strings.TrimSpace(n.CacheDir)
	n.EvictPolicy = strings.TrimSpace(n.EvictPolicy)

	// N1: CORSOrigins 拆分、去空格、去重（保序）
	n.CORSOrigins = normalizeCORSOrigins(n.CORSOrigins)

	// N3: 数值下限（负数→0）
	if n.Port < 0 {
		n.Port = 0
	}
	if n.Workers < 0 {
		n.Workers = 0
	}
	if n.CtxSize < 0 {
		n.CtxSize = 0
	}
	if n.Threads < 0 {
		n.Threads = 0
	}
	if n.MaxTokens < 0 {
		n.MaxTokens = 0
	}
	if n.TimeoutSec < 0 {
		n.TimeoutSec = 0
	}
	if n.MaxConns < 0 {
		n.MaxConns = 0
	}
	if n.Temperature < 0 {
		n.Temperature = 0
	}

	// TopP 不钳制：越界（<0 或 >1）由 Validate V3 报错误
	return n
}

// normalizeCORSOrigins 按逗号拆分 origins，逐项 TrimSpace，去掉空项并按首次出现顺序去重，
// 再以 "," 拼接。空输入返回空串。
func normalizeCORSOrigins(s string) string {
	parts := strings.Split(s, ",")
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return strings.Join(out, ",")
}
