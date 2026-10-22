package config

import "testing"

// Test_NormalizeLlama_when_stringsPadded_trimsAll20TrimmedFields
// 表驱动：20 个 Trim 字段首尾空白在 NormalizeLlama 后全部被去掉。
func Test_NormalizeLlama_when_stringsPadded_trimsAll20TrimmedFields(t *testing.T) {
	in := DefaultLlamaConfig()
	pad := func(s string) string { return " \t" + s + "\n " }

	cases := []struct {
		name string
		set  func(c *LlamaConfig)
		get  func(c LlamaConfig) string
		want string
	}{
		{"StartMode", func(c *LlamaConfig) { c.StartMode = pad("single") }, func(c LlamaConfig) string { return c.StartMode }, "single"},
		{"ModelsPreset", func(c *LlamaConfig) { c.ModelsPreset = pad(`D:\models\preset.json`) }, func(c LlamaConfig) string { return c.ModelsPreset }, `D:\models\preset.json`},
		{"Host", func(c *LlamaConfig) { c.Host = pad("10.0.0.2") }, func(c LlamaConfig) string { return c.Host }, "10.0.0.2"},
		{"ModelPath", func(c *LlamaConfig) { c.ModelPath = pad(`D:\models\qwen.gguf`) }, func(c LlamaConfig) string { return c.ModelPath }, `D:\models\qwen.gguf`},
		{"MmprojPath", func(c *LlamaConfig) { c.MmprojPath = pad(`D:\models\mmproj.bin`) }, func(c LlamaConfig) string { return c.MmprojPath }, `D:\models\mmproj.bin`},
		{"Alias", func(c *LlamaConfig) { c.Alias = pad("qwen") }, func(c LlamaConfig) string { return c.Alias }, "qwen"},
		{"ChatTemplateFile", func(c *LlamaConfig) { c.ChatTemplateFile = pad(`D:\templates\qwen.jinja`) }, func(c LlamaConfig) string { return c.ChatTemplateFile }, `D:\templates\qwen.jinja`},
		{"MoeModelType", func(c *LlamaConfig) { c.MoeModelType = pad("dense") }, func(c LlamaConfig) string { return c.MoeModelType }, "dense"},
		{"MoeCpuMode", func(c *LlamaConfig) { c.MoeCpuMode = pad("off") }, func(c LlamaConfig) string { return c.MoeCpuMode }, "off"},
		{"GpuLayersDraft", func(c *LlamaConfig) { c.GpuLayersDraft = pad("all") }, func(c LlamaConfig) string { return c.GpuLayersDraft }, "all"},
		{"LoadMode", func(c *LlamaConfig) { c.LoadMode = pad("mlock") }, func(c LlamaConfig) string { return c.LoadMode }, "mlock"},
		{"CacheTypeV", func(c *LlamaConfig) { c.CacheTypeV = pad("q8_0") }, func(c LlamaConfig) string { return c.CacheTypeV }, "q8_0"},
		{"CacheTypeK", func(c *LlamaConfig) { c.CacheTypeK = pad("q8_0") }, func(c LlamaConfig) string { return c.CacheTypeK }, "q8_0"},
		{"CacheTypeVDraft", func(c *LlamaConfig) { c.CacheTypeVDraft = pad("q8_0") }, func(c LlamaConfig) string { return c.CacheTypeVDraft }, "q8_0"},
		{"CacheTypeKDraft", func(c *LlamaConfig) { c.CacheTypeKDraft = pad("q8_0") }, func(c LlamaConfig) string { return c.CacheTypeKDraft }, "q8_0"},
		{"SpecType", func(c *LlamaConfig) { c.SpecType = pad("none") }, func(c LlamaConfig) string { return c.SpecType }, "none"},
		{"FlashAttn", func(c *LlamaConfig) { c.FlashAttn = pad("on") }, func(c LlamaConfig) string { return c.FlashAttn }, "on"},
		{"Reasoning", func(c *LlamaConfig) { c.Reasoning = pad("off") }, func(c LlamaConfig) string { return c.Reasoning }, "off"},
		{"ReasoningFormat", func(c *LlamaConfig) { c.ReasoningFormat = pad("deepseek") }, func(c LlamaConfig) string { return c.ReasoningFormat }, "deepseek"},
		{"Fit", func(c *LlamaConfig) { c.Fit = pad("off") }, func(c LlamaConfig) string { return c.Fit }, "off"},
	}

	for _, tc := range cases {
		tc.set(&in)
	}

	got := NormalizeLlama(in)
	for _, tc := range cases {
		if v := tc.get(got); v != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, v, tc.want)
		}
	}
}

// Test_NormalizeLlama_when_chatTemplateKwargsPadded_keepsWhitespace
// ChatTemplateKwargs 属用户文本内容，不 Trim，首尾空白原样保留。
func Test_NormalizeLlama_when_chatTemplateKwargsPadded_keepsWhitespace(t *testing.T) {
	in := DefaultLlamaConfig()
	in.ChatTemplateKwargs = "  {\"max\": 42}  "

	got := NormalizeLlama(in)
	if got.ChatTemplateKwargs != in.ChatTemplateKwargs {
		t.Errorf("ChatTemplateKwargs 不应被 Trim: got %q", got.ChatTemplateKwargs)
	}
}

// Test_NormalizeLlama_when_emptyStrings_stayEmpty 全空字符串配置规整后不变。
func Test_NormalizeLlama_when_emptyStrings_stayEmpty(t *testing.T) {
	in := LlamaConfig{}
	if got := NormalizeLlama(in); got != in {
		t.Errorf("全空串配置规整后应保持不变, got %+v", got)
	}
}

// Test_NormalizeLlama_when_called_doesNotMutateInput 规整不得修改入参。
func Test_NormalizeLlama_when_called_doesNotMutateInput(t *testing.T) {
	in := DefaultLlamaConfig()
	in.ModelPath = " D:\\models\\x.gguf "
	before := in

	_ = NormalizeLlama(in)

	if in != before {
		t.Fatal("NormalizeLlama 修改了入参")
	}
}
