// llamacpp.go llama.cpp 推理引擎的配置结构与持久化。
//
// 与 ninfer 的 34 字段 Config（config.go）完全分离：字段集、默认值、
// argv flag 都不同（见 start-backend.ps1 的 llamacpp 分支）。
// 持久化为推理程序所在目录下的 llama.json（与 config.json 同目录不同名）。
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LlamaConfig llama.cpp 引擎的 18 个配置字段（对应 start-backend.ps1 llamacpp 分支）。
type LlamaConfig struct {
	// 服务
	Host string `json:"host"`
	Port int    `json:"port"`

	// 模型（GGUF 文件路径，-m）
	ModelPath string `json:"modelPath"`

	// 上下文与内存
	Context int `json:"context"` // -c
	Budget  int `json:"budget"`  // --kvmem-budget
	Reserve int `json:"reserve"` // --kvmem-gen-reserve
	HostMiB int `json:"hostMiB"` // --kvmem-host-mib（0=不传）

	// KV 缓存
	KvType       string `json:"kvType"`       // --kv-dtype
	SinkTokens   int    `json:"sinkTokens"`   // --kvmem-sink-tokens
	RecentTokens int    `json:"recentTokens"` // --kvmem-recent-tokens

	// 生成
	MaxTokens int `json:"maxTokens"` // -n
	Prefill   int `json:"prefill"`   // -b 与 -ub 同值

	// LRFU 历史加权
	NoLrfu           bool    `json:"noLrfu"`
	LrfuLambda       float64 `json:"lrfuLambda"`
	LrfuTopFraction  float64 `json:"lrfuTopFraction"`
	ScoreAggregation string  `json:"scoreAggregation"`

	// SSD 分层
	DiskPath string `json:"diskPath"` // 空=关闭
	DiskMiB  int    `json:"diskMiB"`
}

// DefaultLlamaConfig 返回 llama.cpp 引擎的默认值（start-backend.ps1 llamacpp 分支）。
func DefaultLlamaConfig() LlamaConfig {
	return LlamaConfig{
		Host:             "127.0.0.1",
		Port:             18200,
		ModelPath:        "",
		Context:          262144,
		Budget:           32768,
		Reserve:          1024,
		HostMiB:          16384,
		KvType:           "q8_0",
		SinkTokens:       12288,
		RecentTokens:     1024,
		MaxTokens:        1024,
		Prefill:          128,
		NoLrfu:           false,
		LrfuLambda:       0.5,
		LrfuTopFraction:  0.9,
		ScoreAggregation: "max",
		DiskPath:         "",
		DiskMiB:          4096,
	}
}

// LoadLlamaConfig 从 path 读取 llama.json，语义与 LoadConfig 一致：
// 缺文件→默认值；坏 JSON→错误且不写盘；字段缺失→默认值兜底。
func LoadLlamaConfig(path string) (LlamaConfig, error) {
	cfg := DefaultLlamaConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return LlamaConfig{}, fmt.Errorf("读取 llama 配置 %s 失败: %w", path, err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return LlamaConfig{}, fmt.Errorf("解析 llama 配置 %s 失败: %w", path, err)
	}
	return cfg, nil
}

// SaveLlamaConfig 将 cfg 序列化写入 path（llama.json），必要时创建父目录。
func SaveLlamaConfig(cfg LlamaConfig, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 llama 配置失败: %w", err)
	}
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("创建 llama 配置目录 %s 失败: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("写入 llama 配置 %s 失败: %w", path, err)
	}
	return nil
}

// ValidateLlama 校验 llama.cpp 配置，返回错误文案列表（空=通过）。
// 含模型文件存在性检查（fs IO），只可在 goroutine 中调用。
// start-backend.ps1 的 llamacpp 分支不做对齐校验，此处只做基本合法性。
func ValidateLlama(cfg LlamaConfig) []string {
	var errs []string
	if cfg.ModelPath == "" {
		errs = append(errs, "模型路径（-m）不能为空")
	} else if st, err := os.Stat(cfg.ModelPath); err != nil {
		errs = append(errs, "模型文件不存在："+cfg.ModelPath)
	} else if st.IsDir() {
		errs = append(errs, "模型路径是目录，需要 GGUF 文件："+cfg.ModelPath)
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		errs = append(errs, "监听端口必须在 1..65535")
	}
	if cfg.Context <= 0 {
		errs = append(errs, "上下文长度（-c）必须大于 0")
	}
	if cfg.Budget <= 0 {
		errs = append(errs, "KV 预算（--kvmem-budget）必须大于 0")
	}
	if cfg.Reserve <= 0 {
		errs = append(errs, "生成保留（--kvmem-gen-reserve）必须大于 0")
	}
	if cfg.Prefill <= 0 {
		errs = append(errs, "prefill 批大小（-b）必须大于 0")
	}
	if cfg.KvType == "" {
		errs = append(errs, "KV 类型（--kv-dtype）不能为空")
	}
	if cfg.ScoreAggregation != "mean" && cfg.ScoreAggregation != "max" {
		errs = append(errs, "GQA 聚合必须是 mean 或 max")
	}
	if cfg.LrfuLambda < 0 || cfg.LrfuLambda > 1 {
		errs = append(errs, "LRFU lambda 必须在 [0,1]")
	}
	if cfg.LrfuTopFraction <= 0 || cfg.LrfuTopFraction > 1 {
		errs = append(errs, "LRFU top-frac 必须在 (0,1]")
	}
	return errs
}
