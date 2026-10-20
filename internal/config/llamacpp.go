// llamacpp.go llama.cpp 推理引擎的配置结构与持久化。
//
// 与 ninfer 的 32 字段 Config（config_ninfer.go）完全分离：字段集、默认值、
// argv flag 都不同（见 H:\models\Qwen3.8-27B-GSQ_S-MTP.cmd 启动脚本）。
// 持久化为启动器 exe 同级目录下的 llama.json（与 ninfer.json 同目录不同名）。
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LlamaConfig llama.cpp 引擎的 46 个配置字段（对应 Qwen3.8-27B-GSQ_S-MTP.cmd）。
type LlamaConfig struct {
	// 服务
	Host     string `json:"host"`     // --host
	Port     int    `json:"port"`     // --port
	Threads  int    `json:"threads"`  // -t
	Parallel int    `json:"parallel"` // --parallel

	// 模型
	ModelPath        string `json:"modelPath"`        // -m
	MmprojPath       string `json:"mmprojPath"`       // --mmproj（空=不传）
	MmprojOffload    bool   `json:"mmprojOffload"`    // --mmproj-offload
	ImageMinTokens   int    `json:"imageMinTokens"`   // --image-min-tokens
	ImageMaxTokens   int    `json:"imageMaxTokens"`   // --image-max-tokens
	Alias            string `json:"alias"`            // --alias
	ChatTemplateFile string `json:"chatTemplateFile"` // --chat-template-file（空=不传）

	// 上下文与批处理
	CtxSize    int `json:"ctxSize"`    // -c
	NPredict   int `json:"nPredict"`   // -n
	BatchSize  int `json:"batchSize"`  // -b
	UbatchSize int `json:"ubatchSize"` // -ub
	Keep       int `json:"keep"`       // --keep

	// GPU 与加载
	GpuLayers      int    `json:"gpuLayers"`      // --gpu-layers
	GpuLayersDraft string `json:"gpuLayersDraft"` // --gpu-layers-draft
	LoadMode       string `json:"loadMode"`       // --load-mode
	NoWarmup       bool   `json:"noWarmup"`       // --no-warmup

	// KV 缓存
	CacheTypeV       string `json:"cacheTypeV"`       // --cache-type-v
	CacheTypeK       string `json:"cacheTypeK"`       // --cache-type-k
	CacheTypeVDraft  string `json:"cacheTypeVDraft"`  // --cache-type-v-draft
	CacheTypeKDraft  string `json:"cacheTypeKDraft"`  // --cache-type-k-draft
	CacheRam         int    `json:"cacheRam"`         // --cache-ram
	NoCacheIdleSlots bool   `json:"noCacheIdleSlots"` // --no-cache-idle-slots
	NoHost           bool   `json:"noHost"`           // --no-host

	// 推测解码
	SpecType      string `json:"specType"`      // --spec-type
	SpecDraftNMin int    `json:"specDraftNMin"` // --spec-draft-n-min
	SpecDraftNMax int    `json:"specDraftNMax"` // --spec-draft-n-max
	KvUnified     bool   `json:"kvUnified"`     // --kv-unified

	// 注意力与检查点
	FlashAttn      string `json:"flashAttn"`      // -fa（on/off）
	CtxCheckpoints int    `json:"ctxCheckpoints"` // --ctx-checkpoints
	Poll           int    `json:"poll"`           // --poll

	// 推理行为
	Jinja              bool   `json:"jinja"`              // --jinja
	Reasoning          string `json:"reasoning"`          // --reasoning（on/off）
	ReasoningPreserve  bool   `json:"reasoningPreserve"`  // --reasoning-preserve
	ReasoningFormat    string `json:"reasoningFormat"`    // --reasoning-format
	Fit                string `json:"fit"`                // -fit（on/off）
	ChatTemplateKwargs string `json:"chatTemplateKwargs"` // --chat-template-kwargs（空=不传）

	// 采样
	Temp            float64 `json:"temp"`            // --temp
	TopP            float64 `json:"topP"`            // --top-p
	TopK            int     `json:"topK"`            // --top-k
	MinP            float64 `json:"minP"`            // --min-p
	PresencePenalty float64 `json:"presencePenalty"` // --presence-penalty
	RepeatPenalty   float64 `json:"repeatPenalty"`   // --repeat-penalty
}

// DefaultLlamaConfig 返回 llama.cpp 引擎的默认值（Qwen3.8-27B-GSQ_S-MTP.cmd 配方）。
func DefaultLlamaConfig() LlamaConfig {
	return LlamaConfig{
		Host:     "0.0.0.0",
		Port:     8080,
		Threads:  4,
		Parallel: 1,

		ModelPath:        `H:\models\Qwen3.8-GSQ\Qwen3.8-27B-GSQ-RCO-IQ3_S-mtp.gguf`,
		MmprojPath:       `H:\models\Qwen3.8-GSQ\mmproj-Qwen3.8-27B-BF16.gguf`,
		MmprojOffload:    true,
		ImageMinTokens:   1024,
		ImageMaxTokens:   4096,
		Alias:            "Qwen3.8-27B",
		ChatTemplateFile: `Qwen-Sharp-Chat-Templates\chat_template.jinja`,

		CtxSize:    81920,
		NPredict:   16384,
		BatchSize:  512,
		UbatchSize: 512,
		Keep:       -1,

		GpuLayers:      99,
		GpuLayersDraft: "all",
		LoadMode:       "mlock",
		NoWarmup:       true,

		CacheTypeV:       "q4_0",
		CacheTypeK:       "q4_0",
		CacheTypeVDraft:  "q4_0",
		CacheTypeKDraft:  "q4_0",
		CacheRam:         8192,
		NoCacheIdleSlots: true,
		NoHost:           true,

		SpecType:      "draft-mtp",
		SpecDraftNMin: 1,
		SpecDraftNMax: 3,
		KvUnified:     true,

		FlashAttn:      "on",
		CtxCheckpoints: 4,
		Poll:           10,

		Jinja:              true,
		Reasoning:          "on",
		ReasoningPreserve:  true,
		ReasoningFormat:    "deepseek",
		Fit:                "off",
		ChatTemplateKwargs: `{"preserve-thinking": true,"reasoning_effort": "medium"}`,

		Temp:            1,
		TopP:            0.95,
		TopK:            20,
		MinP:            0.0,
		PresencePenalty: 0.0,
		RepeatPenalty:   1.0,
	}
}

// LoadLlamaConfig 从 path 读取 llama.json，语义与 LoadConfig 一致：
// 缺文件→默认值；坏 JSON→错误且不写盘；字段缺失→默认值兜底。
func LoadLlamaConfig(path string) (LlamaConfig, error) {
	cfg := DefaultLlamaConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 缺文件→默认值；但模型/视觉/模板路径不猜默认：为空，等用户指定后写入 json。
			cfg.ModelPath = ""
			cfg.MmprojPath = ""
			cfg.ChatTemplateFile = ""
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
	if cfg.Threads < 0 {
		errs = append(errs, "线程数（-t）不能为负")
	}
	if cfg.Parallel < 1 {
		errs = append(errs, "并行请求数（--parallel）必须 ≥ 1")
	}
	if cfg.CtxSize <= 0 {
		errs = append(errs, "上下文长度（-c）必须大于 0")
	}
	if cfg.NPredict <= 0 {
		errs = append(errs, "最大回复长度（-n）必须大于 0")
	}
	if cfg.BatchSize <= 0 || cfg.UbatchSize <= 0 {
		errs = append(errs, "批大小（-b/-ub）必须大于 0")
	}
	if cfg.UbatchSize > cfg.BatchSize {
		errs = append(errs, "ubatch（-ub）不得超过 batch（-b）")
	}
	if cfg.GpuLayers < 0 {
		errs = append(errs, "GPU 层数（--gpu-layers）不能为负")
	}
	if cfg.CacheRam < 0 {
		errs = append(errs, "缓存内存（--cache-ram）不能为负")
	}
	if cfg.SpecDraftNMin < 0 || cfg.SpecDraftNMax < 0 {
		errs = append(errs, "草稿长度（--spec-draft-n-min/max）不能为负")
	}
	if cfg.SpecDraftNMax > 0 && cfg.SpecDraftNMin > cfg.SpecDraftNMax {
		errs = append(errs, "草稿最小长度不得超过最大长度")
	}
	if cfg.CtxCheckpoints < 0 {
		errs = append(errs, "上下文检查点数（--ctx-checkpoints）不能为负")
	}
	if cfg.Poll < 0 {
		errs = append(errs, "轮询间隔（--poll）不能为负")
	}
	if cfg.ImageMinTokens < 0 || cfg.ImageMaxTokens < 0 {
		errs = append(errs, "视觉 token 数（--image-min/max-tokens）不能为负")
	}
	if cfg.ImageMaxTokens > 0 && cfg.ImageMinTokens > cfg.ImageMaxTokens {
		errs = append(errs, "视觉最小 token 数不得超过最大 token 数")
	}
	if cfg.Temp < 0 {
		errs = append(errs, "温度（--temp）不能为负")
	}
	if cfg.TopP <= 0 || cfg.TopP > 1 {
		errs = append(errs, "top-p 必须在 (0,1]")
	}
	if cfg.TopK < 0 {
		errs = append(errs, "top-k 不能为负")
	}
	if cfg.MinP < 0 || cfg.MinP > 1 {
		errs = append(errs, "min-p 必须在 [0,1]")
	}
	if cfg.RepeatPenalty < 0 {
		errs = append(errs, "重复惩罚（--repeat-penalty）不能为负")
	}
	return errs
}
