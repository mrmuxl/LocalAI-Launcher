package config

import "strings"

// NormalizeLlama 返回一份规整后的 llama 配置（不修改入参），与 ninfer 的
// Normalize 对称：字符串字段 TrimSpace，数值合法性交给 ValidateLlama 报错，
// 不在此钳制。空串/纯空白串处理与 ninfer 一致（统一 TrimSpace，空保持空）。
//
// Trim 的字段（20 个：路径、主机名与枚举/参数值——首尾空白是误输入的噪音，
// 不影响语义且会破坏路径/flag 匹配）：
//   - 路径：ModelsPreset、ModelPath、MmprojPath、ChatTemplateFile
//   - 主机/别名：Host、Alias
//   - 枚举/参数值：StartMode、MoeModelType、MoeCpuMode、GpuLayersDraft、
//     LoadMode、CacheTypeV、CacheTypeK、CacheTypeVDraft、CacheTypeKDraft、
//     SpecType、FlashAttn、Reasoning、ReasoningFormat、Fit
//
// 不 Trim 的字段（1 个：用户文本内容——首尾空白可能承载内容语义，Trim 会
// 改变用户写下的 payload）：
//   - ChatTemplateKwargs：传给聊天模板的参数 JSON 内容，是用户编写的
//     payload 文本，不是路径或枚举值。
func NormalizeLlama(cfg LlamaConfig) LlamaConfig {
	n := cfg
	// 服务/路径
	n.StartMode = strings.TrimSpace(n.StartMode)
	n.ModelsPreset = strings.TrimSpace(n.ModelsPreset)
	n.Host = strings.TrimSpace(n.Host)
	n.ModelPath = strings.TrimSpace(n.ModelPath)
	n.MmprojPath = strings.TrimSpace(n.MmprojPath)
	n.Alias = strings.TrimSpace(n.Alias)
	n.ChatTemplateFile = strings.TrimSpace(n.ChatTemplateFile)
	// 枚举/参数值
	n.MoeModelType = strings.TrimSpace(n.MoeModelType)
	n.MoeCpuMode = strings.TrimSpace(n.MoeCpuMode)
	n.GpuLayersDraft = strings.TrimSpace(n.GpuLayersDraft)
	n.LoadMode = strings.TrimSpace(n.LoadMode)
	n.CacheTypeV = strings.TrimSpace(n.CacheTypeV)
	n.CacheTypeK = strings.TrimSpace(n.CacheTypeK)
	n.CacheTypeVDraft = strings.TrimSpace(n.CacheTypeVDraft)
	n.CacheTypeKDraft = strings.TrimSpace(n.CacheTypeKDraft)
	n.SpecType = strings.TrimSpace(n.SpecType)
	n.FlashAttn = strings.TrimSpace(n.FlashAttn)
	n.Reasoning = strings.TrimSpace(n.Reasoning)
	n.ReasoningFormat = strings.TrimSpace(n.ReasoningFormat)
	n.Fit = strings.TrimSpace(n.Fit)
	// ChatTemplateKwargs 属用户文本内容，不 Trim（理由见头注释）
	return n
}
