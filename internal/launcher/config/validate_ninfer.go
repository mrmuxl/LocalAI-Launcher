package config

import (
	"fmt"
	"os"
	"strings"
)

// 合法枚举取值（V5）。
var (
	validLogLevels     = map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	validPriorities    = map[string]bool{"low": true, "normal": true, "high": true, "realtime": true}
	validEvictPolicies = map[string]bool{"lru": true, "fifo": true}
)

// Validate 对 cfg 执行 V1-V8 语义校验，聚合返回全部错误（不短路）。
// 返回的字符串为"错误码+参数"（形如 "V1:Port=0 out of range [1,65535]"），
// 中文文案由 gui 层按错误码前缀映射。
//
// 说明：V7「合法路径」此处强制为"不含 NUL 字节"。严格的 filepath.Clean 幂等
// 不在此强制——计划自带的默认 CacheDir="./cache" 并非 Clean 幂等（Clean 后会变
// "cache"），若强制会导致默认配置非法，与"默认配置可启动"矛盾。
func Validate(cfg Config) []string {
	var errs []string

	// V1: 端口 / worker 数 / 上下文长度范围
	if cfg.Port < 1 || cfg.Port > 65535 {
		errs = append(errs, fmt.Sprintf("V1:Port=%d out of range [1,65535]", cfg.Port))
	}
	if cfg.Workers < 1 || cfg.Workers > 256 {
		errs = append(errs, fmt.Sprintf("V1:Workers=%d out of range [1,256]", cfg.Workers))
	}
	if cfg.CtxSize < 1 || cfg.CtxSize > 1048576 {
		errs = append(errs, fmt.Sprintf("V1:CtxSize=%d out of range [1,1048576]", cfg.CtxSize))
	}

	// V2: MaxTokens / TimeoutSec / MaxConns
	if cfg.MaxTokens < 1 {
		errs = append(errs, fmt.Sprintf("V2:MaxTokens=%d < 1", cfg.MaxTokens))
	}
	if cfg.TimeoutSec < 1 {
		errs = append(errs, fmt.Sprintf("V2:TimeoutSec=%d < 1", cfg.TimeoutSec))
	}
	if cfg.MaxConns < 1 || cfg.MaxConns > 1024 {
		errs = append(errs, fmt.Sprintf("V2:MaxConns=%d out of range [1,1024]", cfg.MaxConns))
	}

	// V3: Temperature / TopP / TopK / RepeatPenalty
	if cfg.Temperature < 0 {
		errs = append(errs, fmt.Sprintf("V3:Temperature=%g < 0", cfg.Temperature))
	}
	if cfg.TopP < 0 || cfg.TopP > 1 {
		errs = append(errs, fmt.Sprintf("V3:TopP=%g out of range [0,1]", cfg.TopP))
	}
	if cfg.TopK < 0 {
		errs = append(errs, fmt.Sprintf("V3:TopK=%d < 0", cfg.TopK))
	}
	if cfg.RepeatPenalty <= 0 {
		errs = append(errs, fmt.Sprintf("V3:RepeatPenalty=%g <= 0", cfg.RepeatPenalty))
	}

	// V4: TLS 开启时证书/私钥必须非空且文件存在（fs 检查仅在校验阶段）
	if cfg.TLS {
		if cfg.CertPath == "" {
			errs = append(errs, "V4:TLS enabled but CertPath is empty")
		} else if !fileExists(cfg.CertPath) {
			errs = append(errs, fmt.Sprintf("V4:cert file not found: %s", cfg.CertPath))
		}
		if cfg.KeyPath == "" {
			errs = append(errs, "V4:TLS enabled but KeyPath is empty")
		} else if !fileExists(cfg.KeyPath) {
			errs = append(errs, fmt.Sprintf("V4:key file not found: %s", cfg.KeyPath))
		}
	}

	// V5: 枚举字段合法性
	if !validLogLevels[cfg.LogLevel] {
		errs = append(errs, fmt.Sprintf("V5:LogLevel=%q not in {debug,info,warn,error}", cfg.LogLevel))
	}
	if !validPriorities[cfg.Priority] {
		errs = append(errs, fmt.Sprintf("V5:Priority=%q not in {low,normal,high,realtime}", cfg.Priority))
	}
	if !validEvictPolicies[cfg.EvictPolicy] {
		errs = append(errs, fmt.Sprintf("V5:EvictPolicy=%q not in {lru,fifo}", cfg.EvictPolicy))
	}

	// V6: 资源数值
	if cfg.MemLimitGB < 0 {
		errs = append(errs, fmt.Sprintf("V6:MemLimitGB=%d < 0", cfg.MemLimitGB))
	}
	if cfg.SwapGB < 0 {
		errs = append(errs, fmt.Sprintf("V6:SwapGB=%d < 0", cfg.SwapGB))
	}
	if cfg.NumaNode != -1 && cfg.NumaNode < 0 {
		errs = append(errs, fmt.Sprintf("V6:NumaNode=%d must be -1 or >= 0", cfg.NumaNode))
	}
	if cfg.CacheSizeMB < 1 {
		errs = append(errs, fmt.Sprintf("V6:CacheSizeMB=%d < 1", cfg.CacheSizeMB))
	}

	// V7: 路径类字段如非空必须为合法路径（不含 NUL 字节）
	for _, p := range []string{cfg.ModelPath, cfg.CacheDir, cfg.LogFile, cfg.CertPath, cfg.KeyPath} {
		if p != "" && strings.ContainsRune(p, 0) {
			errs = append(errs, "V7:path contains NUL byte")
		}
	}

	// V8: 已在上方聚合全部错误（不短路），此处无额外动作。
	return errs
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
