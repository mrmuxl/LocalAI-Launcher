package config

// DefaultConfig 返回全部 32 字段的默认值（start-ninfer.bat 默认配方）。
func DefaultConfig() Config {
	return Config{
		// 服务
		ListenHost:  DefListenHost,
		Port:        DefPort,
		Concurrency: DefConcurrency,

		// 模型
		ModelPath:   DefModelPath,
		Gpu:         DefGpu,
		StageLayers: DefStageLayers,

		// 上下文与内存
		Context:   DefContext,
		Budget:    DefBudget,
		Reserve:   DefReserve,
		HostMiB:   DefHostMiB,
		MaxTokens: DefMaxTokens,
		Prefill:   DefPrefill,

		// KV 缓存
		KvType:           DefKvType,
		SinkTokens:       DefSinkTokens,
		RecentTokens:     DefRecentTokens,
		RetainedSessions: DefRetainedSessions,
		FastPrefill:      DefFastPrefill,

		// 推测解码
		MtpDrafts:     DefMtpDrafts,
		NgramDrafts:   DefNgramDrafts,
		NgramMinMatch: DefNgramMinMatch,
		AdaptiveMtp:   DefAdaptiveMtp,

		// 视觉
		Vision:          DefVision,
		VisionResidency: DefVisionResidency,
		VisionTokens:    DefVisionTokens,

		// 思考与设备
		ThinkingBudget:         DefThinkingBudget,
		DefaultReasoningEffort: "", // ps1 未传，空 = 不传
		DeviceProfile:          DefDeviceProfile,

		// SSD 分层
		DiskPath: DefDiskPath,
		DiskMiB:  DefDiskMiB,

		// 检索调优
		NoLrfu:           DefNoLrfu,
		LrfuLambda:       DefLrfuLambda,
		LrfuTopFraction:  DefLrfuTopFraction,
		ScoreAggregation: DefScoreAggregation,
	}
}
