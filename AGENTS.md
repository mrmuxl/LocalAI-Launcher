# AGENTS.md — NinferManager

> 本文件是给 AI 编码代理的项目级硬性规则说明。

## 线程原则（硬规则）

- 凡是不能立即返回的按钮、页面，都必须使用 goroutine 在后台处理，渲染更新一律回到 UI 线程
- 事件处理器（Tappable、OnChanged 等）里禁止直接执行网络请求、文件 IO、数据库访问、`time.Sleep` 等耗时操作
- 耗时逻辑放入 `go func() { ... }()` 或等价后台 goroutine；UI 操作回到主线程用 `fyne.Do(func() { ... })` 调度
- 禁止跨 goroutine 直接读写 Fyne 控件（Label、Entry、List、Widget 树等），必须经 `fyne.Do` 回主线程
- goroutine 内的错误不得静默丢弃：至少要 `log` 或经 UI 线程反馈给用户

## Fyne 专项细则

- Fyne 控件回调（`Tapped`、`OnChanged` 等）运行在主线程，必须立即返回，任何阻塞逻辑一律放入后台 goroutine
- 一切 UI 更新（控件读写）必须经 `fyne.Do(func() { ... })` 调度回主线程，非主线程禁止直接读写 Fyne 组件
- `fyne.NewApp()` 与 `app.Run()` 只在主 goroutine 调用一次，业务代码禁止自行启动或退出 Fyne 事件循环
- 按钮执行耗时操作时，点击后立即 `Disable()` 防止重复触发，后台任务完成后经 `fyne.Do` 调用 `Enable()` 恢复
- 不能立即渲染的页面，先在主线程显示 loading/骨架态，后台数据就绪后经 `fyne.Do` 异步替换为实际内容
- 动画帧回调（`canvas.Animation` 的 `OnTick` 等）运行在主线程，帧内只允许轻量绘制，重计算一律放入 goroutine，禁止在非主线程启动/停止动画

## Go 参考模式

以下为「按钮 Tap → context goroutine（后台工作）→ 主线程更新（`fyne.Do`）→ 按钮禁用/恢复」的参考实现，仓库中所有按钮/页面代码一律遵循此模式。

```go
package gui

import (
	"context"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// setupButton 展示参考模式：主线程回调中先禁用按钮，耗时逻辑放入
// 后台 goroutine，UI 更新一律经 fyne.Do 回到主线程。
func setupButton(btn *widget.Button, label *widget.Label, work func(context.Context) error) {
	btn.OnTapped = func() {
		// 主线程回调：立即禁用按钮，防止重复触发
		btn.Disable()

		ctx, cancel := context.WithCancel(context.Background())

		// 耗时逻辑放入后台 goroutine
		go func() {
			defer cancel()
			if err := work(ctx); err != nil {
				// goroutine 内的错误不得静默丢弃
				log.Printf("后台任务失败: %v", err)
			}

			// 一切 UI 更新必须经 fyne.Do 调度回主线程
			fyne.Do(func() {
				label.SetText("完成")
				btn.Enable()
			})
		}()
	}
}
```
