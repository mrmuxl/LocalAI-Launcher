package version

import (
	"strings"
	"testing"
)

// TestVersionConst 断言 Version 常量为预期值。
func TestVersionConst(t *testing.T) {
	if Version != "0.1.0" {
		t.Errorf("Version = %q, want %q", Version, "0.1.0")
	}
}

// TestBanner 断言 Banner 返回固定的版本横幅。
func TestBanner(t *testing.T) {
	got := Banner()
	want := "ninfer-launcher v0.1.0"
	if got != want {
		t.Errorf("Banner() = %q, want %q", got, want)
	}
}

// TestBannerContainsVersion 断言 Banner 包含 Version 常量值，
// 防止将来常量改动后 Banner 不同步。
func TestBannerContainsVersion(t *testing.T) {
	if !strings.Contains(Banner(), Version) {
		t.Errorf("Banner() = %q does not contain Version = %q", Banner(), Version)
	}
}
