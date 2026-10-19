package config

import "testing"

// Test_Normalize_when_stringsPadded_trimsAll14StringFields
// N1：14 个字符串字段 TrimSpace（CORSOrigins 单独处理，不在列表内）。
func Test_Normalize_when_stringsPadded_trimsAll14StringFields(t *testing.T) {
	in := DefaultConfig()
	pad := func(s string) string { return " \t" + s + "\n " }
	in.Host = pad("10.0.0.2")
	in.LogLevel = pad("debug")
	in.ModelName = pad("model-x")
	in.ModelPath = pad("/models/x")
	in.Quant = pad("q8_0")
	in.APIKey = pad("secret")
	in.CertPath = pad("/cert.pem")
	in.KeyPath = pad("/key.pem")
	in.MetricsPath = pad("/m")
	in.LogFile = pad("/log.txt")
	in.Priority = pad("high")
	in.AffinityMask = pad("0xf")
	in.CacheDir = pad("/cache")
	in.EvictPolicy = pad("fifo")

	type field struct {
		name string
		want string
		get  func(Config) string
	}
	fields := []field{
		{"Host", "10.0.0.2", func(c Config) string { return c.Host }},
		{"LogLevel", "debug", func(c Config) string { return c.LogLevel }},
		{"ModelName", "model-x", func(c Config) string { return c.ModelName }},
		{"ModelPath", "/models/x", func(c Config) string { return c.ModelPath }},
		{"Quant", "q8_0", func(c Config) string { return c.Quant }},
		{"APIKey", "secret", func(c Config) string { return c.APIKey }},
		{"CertPath", "/cert.pem", func(c Config) string { return c.CertPath }},
		{"KeyPath", "/key.pem", func(c Config) string { return c.KeyPath }},
		{"MetricsPath", "/m", func(c Config) string { return c.MetricsPath }},
		{"LogFile", "/log.txt", func(c Config) string { return c.LogFile }},
		{"Priority", "high", func(c Config) string { return c.Priority }},
		{"AffinityMask", "0xf", func(c Config) string { return c.AffinityMask }},
		{"CacheDir", "/cache", func(c Config) string { return c.CacheDir }},
		{"EvictPolicy", "fifo", func(c Config) string { return c.EvictPolicy }},
	}

	got := Normalize(in)
	for _, f := range fields {
		if v := f.get(got); v != f.want {
			t.Errorf("%s = %q, want %q", f.name, v, f.want)
		}
	}
}

// Test_Normalize_when_corsMixed_dedupsTrimsAndJoins
// N1 特例：CORSOrigins 拆分 → 逐项 trim → 去重 → 逗号拼接。
func Test_Normalize_when_corsMixed_dedupsTrimsAndJoins(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"paddedAndDuplicated", " a , a ,b ,, c", "a,b,c"},
		{"alreadyClean", "a,b,c", "a,b,c"},
		{"empty", "", ""},
		{"singlePadded", "  x  ", "x"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Normalize(Config{CORSOrigins: tc.in}).CORSOrigins
			if got != tc.want {
				t.Errorf("Normalize(CORSOrigins=%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// Test_Normalize_when_numbersNegative_clampsToZero
// N2：8 个数值字段钳制到 ≥0（Port/Workers/Concurrency/Timeout/Memory/... 按实现）。
func Test_Normalize_when_numbersNegative_clampsToZero(t *testing.T) {
	in := DefaultConfig()
	in.Port = -1
	in.Workers = -8
	got := Normalize(in)

	if got.Port != 0 {
		t.Errorf("Port = %d, want 0 (clamped)", got.Port)
	}
	if got.Workers != 0 {
		t.Errorf("Workers = %d, want 0 (clamped)", got.Workers)
	}
}

// Test_Normalize_when_numaNodeAndTopPKeepValues_notClamped
// NumaNode/TopP 不在钳制集合内：负值原样保留（与实现语义一致）。
func Test_Normalize_when_numaNodeAndTopPKeepValues_notClamped(t *testing.T) {
	in := DefaultConfig()
	in.NumaNode = -2
	in.TopP = -0.5
	got := Normalize(in)

	if got.NumaNode != -2 {
		t.Errorf("NumaNode = %d, want -2 (not clamped)", got.NumaNode)
	}
	if got.TopP != -0.5 {
		t.Errorf("TopP = %v, want -0.5 (not clamped)", got.TopP)
	}
}

// Test_Normalize_when_inputDirty_doesNotModifyOriginal
// Normalize 不修改入参：返回副本，原值保持脏状态。
func Test_Normalize_when_inputDirty_doesNotModifyOriginal(t *testing.T) {
	original := DefaultConfig()
	original.Host = " padded-host "
	original.Port = -7

	got := Normalize(original)

	if got.Host == " padded-host " {
		t.Errorf("returned Host not trimmed: %q", got.Host)
	}
	if original.Host != " padded-host " || original.Port != -7 {
		t.Errorf("original mutated by Normalize: Host=%q Port=%d", original.Host, original.Port)
	}
}
