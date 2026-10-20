package config

// DefaultConfig 返回计划「字段表」中全部 34 字段的默认值。
// 空字符串字段（ModelPath/APIKey/CertPath/KeyPath/LogFile/AffinityMask）保持空。
func DefaultConfig() Config {
	return Config{
		// 服务
		Host:     DefHost,
		Port:     DefPort,
		LogLevel: DefLogLevel,
		Workers:  DefWorkers,

		// 模型
		ModelName: DefModelName,
		ModelPath: "", // 空=自动
		CtxSize:   DefCtxSize,
		Threads:   DefThreads, // 0=自动
		Quant:     DefQuant,

		// 采样
		Temperature:   DefTemperature,
		TopP:          DefTopP,
		TopK:          DefTopK,
		RepeatPenalty: DefRepeatPenalty,
		MaxTokens:     DefMaxTokens,

		// 网络
		APIKey:      "", // 可选
		CORSOrigins: DefCORSOrigins,
		TimeoutSec:  DefTimeoutSec,
		MaxConns:    DefMaxConns,

		// 安全
		TLS:      false,
		CertPath: "",
		KeyPath:  "",

		// 观测
		MetricsEnabled: DefMetricsEnabled,
		MetricsPath:    DefMetricsPath,
		LogFile:        "", // 空=stdout
		AccessLog:      DefAccessLog,

		// 资源
		MemLimitGB:   0, // 0=不限
		SwapGB:       0,
		NumaNode:     DefNumaNode, // -1=自动
		Priority:     DefPriority,
		AffinityMask: "",

		// 缓存
		CacheDir:    DefCacheDir,
		CacheSizeMB: DefCacheSizeMB,
		EvictPolicy: DefEvictPolicy,
		Prefetch:    false,
	}
}
