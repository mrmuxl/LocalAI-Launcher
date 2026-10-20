//go:build !windows

// hidewindow_other.go 非 Windows 平台无控制台窗口问题，空实现。
package process

import "os/exec"

// hideConsoleWindow 非 Windows 平台为空实现（保持跨平台调用点统一）。
func hideConsoleWindow(*exec.Cmd) {}
