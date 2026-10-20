package argv

import (
	"slices"
	"testing"

	"llmlauncher/internal/config"
)

// TestBuildArgs_batRecipe 用表驱动用例锁定真实命令行契约
// （对齐 start-backend.ps1 ninfer 分支）：默认配方 + 变体，逐元素精确相等。
func TestBuildArgs_batRecipe(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		want []string
	}{
		{
			name: "bat 默认配方",
			cfg:  config.DefaultConfig(),
			want: []string{
				"Qwen3.8-27B-GSQ-RCO-IQ3_S-ninfer-v3.ninfer",
				"--host", "127.0.0.1",
				"--port", "8080",
				"--max-context", "131072",
				"--max-concurrency", "1",
				"--prefill-chunk", "512",
				"--default-max-tokens", "16384",
				"--kvmem-budget", "36864",
				"--kvmem-gen-reserve", "16384",
				"--kvmem-host-mib", "12288",
				"--kvmem-sink-tokens", "12288",
				"--kvmem-recent-tokens", "1024",
				"--kvmem-sessions", "1",
				"--kv-dtype", "int8",
				"--device-profile", "off",
				"--device", "0",
				"--spec", "mtp",
				"--draft-tokens", "4",
				"--ngram-draft-tokens", "31",
				"--ngram-min-match", "12",
				"--adaptive-mtp",
				"--vision",
				"--vision-residency", "cpu",
				"--vision-max-merged", "1024",
				"--default-thinking-budget", "8192",
				"--kvmem-disk-path", `D:\kv-cache`,
				"--kvmem-disk-mib", "4096",
				"--kvmem-lrfu",
				"--kvmem-lrfu-lambda", "0.5",
				"--kvmem-lrfu-top-frac", "0.9",
				"--kvmem-score-agg", "max",
			},
		},
		{
			name: "MTP 关闭：无 spec 段；NoLrfu：无 LRFU 段；DiskPath 空：无 SSD 段",
			cfg: func() config.Config {
				cfg := config.DefaultConfig()
				cfg.MtpDrafts = 0
				cfg.AdaptiveMtp = false
				cfg.NoLrfu = true
				cfg.DiskPath = ""
				cfg.ThinkingBudget = 0
				cfg.FastPrefill = true
				cfg.Vision = false
				cfg.StageLayers = "24,24"
				return cfg
			}(),
			want: []string{
				"Qwen3.8-27B-GSQ-RCO-IQ3_S-ninfer-v3.ninfer",
				"--host", "127.0.0.1",
				"--port", "8080",
				"--max-context", "131072",
				"--max-concurrency", "1",
				"--prefill-chunk", "512",
				"--default-max-tokens", "16384",
				"--kvmem-budget", "36864",
				"--kvmem-gen-reserve", "16384",
				"--kvmem-host-mib", "12288",
				"--kvmem-sink-tokens", "12288",
				"--kvmem-recent-tokens", "1024",
				"--kvmem-sessions", "1",
				"--kv-dtype", "int8",
				"--device-profile", "off",
				"--devices", "0,1",
				"--stage-layers", "24,24",
				"--fast-prefill-kernel",
				"--kvmem-score-agg", "max",
			},
		},
		{
			name: "思考力度：DefaultReasoningEffort 非空 → --default-reasoning-effort",
			cfg: func() config.Config {
				cfg := config.DefaultConfig()
				cfg.DefaultReasoningEffort = "high"
				return cfg
			}(),
			want: []string{
				"Qwen3.8-27B-GSQ-RCO-IQ3_S-ninfer-v3.ninfer",
				"--host", "127.0.0.1",
				"--port", "8080",
				"--max-context", "131072",
				"--max-concurrency", "1",
				"--prefill-chunk", "512",
				"--default-max-tokens", "16384",
				"--kvmem-budget", "36864",
				"--kvmem-gen-reserve", "16384",
				"--kvmem-host-mib", "12288",
				"--kvmem-sink-tokens", "12288",
				"--kvmem-recent-tokens", "1024",
				"--kvmem-sessions", "1",
				"--kv-dtype", "int8",
				"--device-profile", "off",
				"--device", "0",
				"--spec", "mtp",
				"--draft-tokens", "4",
				"--ngram-draft-tokens", "31",
				"--ngram-min-match", "12",
				"--adaptive-mtp",
				"--vision",
				"--vision-residency", "cpu",
				"--vision-max-merged", "1024",
				"--default-thinking-budget", "8192",
				"--default-reasoning-effort", "high",
				"--kvmem-disk-path", `D:\kv-cache`,
				"--kvmem-disk-mib", "4096",
				"--kvmem-lrfu",
				"--kvmem-lrfu-lambda", "0.5",
				"--kvmem-lrfu-top-frac", "0.9",
				"--kvmem-score-agg", "max",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildArgs(tt.cfg)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("argv mismatch:\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}
