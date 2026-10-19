// main.go 是启动器入口：单实例命名互斥锁 → 定位 worker → 构建 GUI → 进入事件循环。
//
// 线程模型（AGENTS.md）：主 goroutine 只做互斥锁、GUI 构建与事件循环；
// 启动/停止/保存/加载/校验等耗时逻辑全部在 gui/process 包的后台 goroutine 内。
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"ninferrmanager/internal/launcher/gui"
	"ninferrmanager/internal/launcher/process"
	"ninferrmanager/internal/launcher/version"
)

func main() {
	// 单实例锁：Windows 为 Global\NinferLauncher 命名互斥锁，其他平台 no-op。
	if err := acquireSingleInstanceLock(); err == errAlreadyRunning {
		showSecondInstanceDialog()
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, version.Banner(), err)
		os.Exit(1)
	}

	fmt.Println(version.Banner())

	// NewApp 必须在主 goroutine 调用（fyne.NewApp 约束）。
	app := gui.NewApp(resolveWorkerPath())
	app.Run()
}

// resolveWorkerPath 定位与 launcher exe 同目录的 worker 可执行文件（不依赖 CWD）；
// 无法获取 exe 路径时回退 process.DefaultWorkerPath。
func resolveWorkerPath() string {
	exe, err := os.Executable()
	if err != nil {
		return process.DefaultWorkerPath
	}
	return filepath.Join(filepath.Dir(exe), "ninfer.exe")
}

