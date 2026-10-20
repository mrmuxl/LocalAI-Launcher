//go:build windows

// hidewindow_windows.go 隐藏子进程控制台窗口：推理引擎是控制台程序，
// GUI 父进程直接 Start 会弹出黑色 cmd 窗口；CREATE_NO_WINDOW 抑制之。
package process

import (
	"os/exec"
	"syscall"
)

// creationNoWindow 为 Windows CREATE_NO_WINDOW 标志（0x08000000）：
// 子进程不创建控制台窗口，stdout/stderr 管道照常工作。
const creationNoWindow = 0x08000000

// hideConsoleWindow 隐藏子进程的控制台窗口（仅 Windows 生效）。
func hideConsoleWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags |= creationNoWindow
}
