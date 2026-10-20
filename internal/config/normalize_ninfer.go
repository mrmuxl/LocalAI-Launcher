package config

import "strings"

// Normalize 返回一份规整后的配置（不修改入参）：全部字符串字段 TrimSpace。
// 数值合法性（范围/对齐/组合约束）一律交给 Validate 报错，不在此钳制。
func Normalize(cfg Config) Config {
	n := cfg
	n.ListenHost = strings.TrimSpace(n.ListenHost)
	n.ModelPath = strings.TrimSpace(n.ModelPath)
	n.Gpu = strings.TrimSpace(n.Gpu)
	n.StageLayers = strings.TrimSpace(n.StageLayers)
	n.KvType = strings.TrimSpace(n.KvType)
	n.VisionResidency = strings.TrimSpace(n.VisionResidency)
	n.DeviceProfile = strings.TrimSpace(n.DeviceProfile)
	n.DiskPath = strings.TrimSpace(n.DiskPath)
	n.ScoreAggregation = strings.TrimSpace(n.ScoreAggregation)
	return n
}
