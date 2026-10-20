// Package config 定义 NInfer 启动器的配置模型、默认值、JSON 读写、
// 规范化（Normalize）与校验（Validate）。
//
// 本包仅依赖标准库；错误码+参数由本包返回，中文文案在 gui 层组装。
//
// Config 共 34 字段 / 8 分组，与计划「字段表」一一对应（不可增减），
// 默认值对照 docs/scripts/windows/start-ninfer.ps1 对应行。
package config

// Config 是启动器持久化到 config.json 的完整配置。
type Config struct {
	// ── 服务（4）────────────────────────────
	Host     string `json:"host"`     // 1：监听地址
	Port     int    `json:"port"`     // 2：监听端口
	LogLevel string `json:"logLevel"` // 3：debug/info/warn/error
	Workers  int    `json:"workers"`  // 4：并发 worker 数

	// ── 模型（5）────────────────────────────
	ModelName string `json:"modelName"` // 5：模型标识
	ModelPath string `json:"modelPath"` // 6：模型文件/目录，空=自动
	CtxSize   int    `json:"ctxSize"`   // 7：上下文长度
	Threads   int    `json:"threads"`   // 8：推理线程，0=自动
	Quant     string `json:"quant"`     // 9：量化格式

	// ── 采样（5）────────────────────────────
	Temperature   float64 `json:"temperature"`   // 10：采样温度
	TopP          float64 `json:"topP"`          // 11：top-p
	TopK          int     `json:"topK"`          // 12：top-k
	RepeatPenalty float64 `json:"repeatPenalty"` // 13：重复惩罚
	MaxTokens     int     `json:"maxTokens"`     // 14：最大生成 token

	// ── 网络（4）────────────────────────────
	APIKey      string `json:"apiKey"`      // 15：可选鉴权
	CORSOrigins string `json:"corsOrigins"` // 16：允许来源，逗号分隔
	TimeoutSec  int    `json:"timeoutSec"`  // 17：请求超时
	MaxConns    int    `json:"maxConns"`    // 18：最大并发连接

	// ── 安全（3）────────────────────────────
	TLS      bool   `json:"tls"`      // 19：是否启用 TLS
	CertPath string `json:"certPath"` // 20：证书路径
	KeyPath  string `json:"keyPath"`  // 21：私钥路径

	// ── 观测（4）────────────────────────────
	MetricsEnabled bool   `json:"metricsEnabled"` // 22：/metrics 开关
	MetricsPath    string `json:"metricsPath"`    // 23：指标路径
	LogFile        string `json:"logFile"`        // 24：日志文件，空=stdout
	AccessLog      bool   `json:"accessLog"`      // 25：访问日志

	// ── 资源（5）────────────────────────────
	MemLimitGB   int    `json:"memLimitGB"`   // 26：内存上限 GB，0=不限
	SwapGB       int    `json:"swapGB"`       // 27：swap 上限 GB
	NumaNode     int    `json:"numaNode"`     // 28：NUMA 节点，-1=自动
	Priority     string `json:"priority"`     // 29：进程优先级
	AffinityMask string `json:"affinityMask"` // 30：CPU 亲和掩码，空=默认

	// ── 缓存（4）────────────────────────────
	CacheDir    string `json:"cacheDir"`    // 31：缓存目录
	CacheSizeMB int    `json:"cacheSizeMB"` // 32：缓存大小 MB
	EvictPolicy string `json:"evictPolicy"` // 33：lru/fifo
	Prefetch    bool   `json:"prefetch"`    // 34：预取开关
}

// 默认值常量（对照计划「字段表」默认列）。
const (
	DefHost           = "127.0.0.1"
	DefPort           = 8080
	DefLogLevel       = "info"
	DefWorkers        = 4
	DefModelName      = "qwen3-27b"
	DefCtxSize        = 32768
	DefThreads        = 0 // 自动
	DefQuant          = "q4_k_m"
	DefTemperature    = 0.7
	DefTopP           = 0.9
	DefTopK           = 40
	DefRepeatPenalty  = 1.1
	DefMaxTokens      = 4096
	DefCORSOrigins    = "*"
	DefTimeoutSec     = 300
	DefMaxConns       = 32
	DefMetricsEnabled = true
	DefMetricsPath    = "/metrics"
	DefAccessLog      = true
	DefNumaNode       = -1 // 自动
	DefPriority       = "normal"
	DefCacheDir       = "./cache"
	DefCacheSizeMB    = 1024
	DefEvictPolicy    = "lru"
)
