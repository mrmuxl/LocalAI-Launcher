// main.go 是启动器入口：单实例命名互斥锁 → 定位 worker → 构建 GUI → 进入事件循环。
//
// 线程模型（AGENTS.md）：主 goroutine 只做互斥锁、GUI 构建与事件循环；
// 启动/停止/保存/加载/校验等耗时逻辑全部在 gui/process 包的后台 goroutine 内。
package main

import (
	"fmt"
	"os"

	"ninferrmanager/internal/launcher/gui"
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

	// NewApp 必须在主 goroutine 调用（app.New 约束）；
	// worker 路径由 App 启动后经存档或文件选择框确定。
	app := gui.NewApp()
	app.Run()
}
