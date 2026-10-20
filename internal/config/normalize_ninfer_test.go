package config

import "testing"

// Test_Normalize_when_stringsPadded_trimsAll9StringFields
// 全部 9 个字符串字段 TrimSpace。
func Test_Normalize_when_stringsPadded_trimsAll9StringFields(t *testing.T) {
	in := DefaultConfig()
	pad := func(s string) string { return " \t" + s + "\n " }
	in.ListenHost = pad("10.0.0.2")
	in.ModelPath = pad(`D:\models\x.ninfer`)
	in.Gpu = pad("0,1")
	in.StageLayers = pad("24,24")
	in.KvType = pad("int8")
	in.VisionResidency = pad("cpu")
	in.DeviceProfile = pad("auto")
	in.DiskPath = pad(`D:\kv-cache`)
	in.ScoreAggregation = pad("max")

	got := Normalize(in)

	if got.ListenHost != "10.0.0.2" {
		t.Errorf("ListenHost = %q", got.ListenHost)
	}
	if got.ModelPath != `D:\models\x.ninfer` {
		t.Errorf("ModelPath = %q", got.ModelPath)
	}
	if got.Gpu != "0,1" {
		t.Errorf("Gpu = %q", got.Gpu)
	}
	if got.StageLayers != "24,24" {
		t.Errorf("StageLayers = %q", got.StageLayers)
	}
	if got.KvType != "int8" {
		t.Errorf("KvType = %q", got.KvType)
	}
	if got.VisionResidency != "cpu" {
		t.Errorf("VisionResidency = %q", got.VisionResidency)
	}
	if got.DeviceProfile != "auto" {
		t.Errorf("DeviceProfile = %q", got.DeviceProfile)
	}
	if got.DiskPath != `D:\kv-cache` {
		t.Errorf("DiskPath = %q", got.DiskPath)
	}
	if got.ScoreAggregation != "max" {
		t.Errorf("ScoreAggregation = %q", got.ScoreAggregation)
	}
}

// Test_Normalize_when_called_doesNotMutateInput 规整不得修改入参。
func Test_Normalize_when_called_doesNotMutateInput(t *testing.T) {
	in := DefaultConfig()
	in.ListenHost = " 127.0.0.1 "
	before := in
	_ = Normalize(in)
	if in != before {
		t.Fatal("Normalize 修改了入参")
	}
}
