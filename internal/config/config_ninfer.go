// Package config 定义 NInfer 启动器的配置模型、默认值、JSON 读写、
// 规范化（Normalize）与校验（Validate）。
//
// 本包仅依赖标准库；Validate 直接返回中文错误文案（与 ValidateLlama 一致）。
//
// Config 共 32 字段 / 9 分组，与 ninfer-kvmem-server.exe 的真实命令行参数
// 一一对应（来源：start-ninfer.bat 默认配方 → start-ninfer.ps1 →
// start-backend.ps1 ninfer 分支），默认值即该配方的值。
package config

// Config 是启动器持久化到 config.json 的完整配置。
type Config struct {
	// ── 服务（3）────────────────────────────
	ListenHost  string `json:"listenHost"`  // 1：监听地址（--host）
	Port        int    `json:"port"`        // 2：监听端口（--port）
	Concurrency int    `json:"concurrency"` // 3：并发解码请求数（--max-concurrency）

	// ── 模型（3）────────────────────────────
	ModelPath   string `json:"modelPath"`   // 4：模型文件（位置参数 model.ninfer）
	Gpu         string `json:"gpu"`         // 5：GPU（CUDA_VISIBLE_DEVICES 环境变量）
	StageLayers string `json:"stageLayers"` // 6：每卡层数（--stage-layers，空=不传）

	// ── 上下文与内存（6）────────────────────
	Context   int `json:"context"`   // 7：逻辑上下文上限（--max-context）
	Budget    int `json:"budget"`    // 8：历史预算 B（--kvmem-budget）
	Reserve   int `json:"reserve"`   // 9：生成保留 R（--kvmem-gen-reserve）
	HostMiB   int `json:"hostMiB"`   // 10：Host KV 预算（--kvmem-host-mib）
	MaxTokens int `json:"maxTokens"` // 11：默认输出上限（--default-max-tokens）
	Prefill   int `json:"prefill"`   // 12：prefill 块（--prefill-chunk）

	// ── KV 缓存（5）────────────────────────
	KvType           string `json:"kvType"`           // 13：KV 存储（--kv-dtype）
	SinkTokens       int    `json:"sinkTokens"`       // 14：常驻前缀（--kvmem-sink-tokens）
	RecentTokens     int    `json:"recentTokens"`     // 15：常驻后缀（--kvmem-recent-tokens）
	RetainedSessions int    `json:"retainedSessions"` // 16：保留历史数（--kvmem-sessions）
	FastPrefill      bool   `json:"fastPrefill"`      // 17：快速 prefill 内核（--fast-prefill-kernel）

	// ── 推测解码（4）────────────────────────
	MtpDrafts     int  `json:"mtpDrafts"`     // 18：MTP 草稿数（--spec mtp --draft-tokens，0=关闭）
	NgramDrafts   int  `json:"ngramDrafts"`   // 19：ngram 草稿数（--ngram-draft-tokens）
	NgramMinMatch int  `json:"ngramMinMatch"` // 20：最短匹配（--ngram-min-match）
	AdaptiveMtp   bool `json:"adaptiveMtp"`   // 21：自适应 MTP（--adaptive-mtp）

	// ── 视觉（3）──────────────────────────
	Vision          bool   `json:"vision"`          // 22：启用视觉（--vision）
	VisionResidency string `json:"visionResidency"` // 23：视觉驻留（--vision-residency）
	VisionTokens    int    `json:"visionTokens"`    // 24：单媒体合并上限（--vision-max-merged）

	// ── 思考与设备（2）────────────────────
	ThinkingBudget int    `json:"thinkingBudget"` // 25：默认思考预算（--default-thinking-budget，0=不传）
	DeviceProfile  string `json:"deviceProfile"`  // 26：设备档案（--device-profile）

	// ── SSD 分层（2）──────────────────────
	DiskPath string `json:"diskPath"` // 27：SSD 分层目录（--kvmem-disk-path，空=不传）
	DiskMiB  int    `json:"diskMiB"`  // 28：SSD 配额（--kvmem-disk-mib）

	// ── 检索调优（4）──────────────────────
	NoLrfu           bool    `json:"noLrfu"`           // 29：关闭 LRFU（!→--kvmem-lrfu）
	LrfuLambda       float64 `json:"lrfuLambda"`       // 30：历史衰减（--kvmem-lrfu-lambda）
	LrfuTopFraction  float64 `json:"lrfuTopFraction"`  // 31：命中比例（--kvmem-lrfu-top-frac）
	ScoreAggregation string  `json:"scoreAggregation"` // 32：评分聚合（--kvmem-score-agg）
}

// 默认值常量（来源：start-ninfer.bat 默认配方 + start-ninfer.ps1 默认值）。
const (
	DefListenHost  = "127.0.0.1"
	DefPort        = 8080
	DefConcurrency = 1

	DefModelPath   = "Qwen3.8-27B-GSQ-RCO-IQ3_S-ninfer-v3.ninfer"
	DefGpu         = "0"
	DefStageLayers = "" // 空=不传

	DefContext   = 131072
	DefBudget    = 36864
	DefReserve   = 16384
	DefHostMiB   = 12288
	DefMaxTokens = 16384
	DefPrefill   = 512

	DefKvType           = "int8"
	DefSinkTokens       = 12288
	DefRecentTokens     = 1024
	DefRetainedSessions = 1
	DefFastPrefill      = false

	DefMtpDrafts     = 4
	DefNgramDrafts   = 31
	DefNgramMinMatch = 12
	DefAdaptiveMtp   = true

	DefVision          = true
	DefVisionResidency = "cpu"
	DefVisionTokens    = 1024

	DefThinkingBudget = 8192
	DefDeviceProfile  = "off"

	DefDiskPath = `D:\kv-cache`
	DefDiskMiB  = 4096

	DefNoLrfu           = false
	DefLrfuLambda       = 0.5
	DefLrfuTopFraction  = 0.9
	DefScoreAggregation = "max"
)
