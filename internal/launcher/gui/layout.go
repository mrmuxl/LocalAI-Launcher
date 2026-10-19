// layout.go 组装窗口布局：顶部（保存/重新加载/启动/停止操作行）/ 两栏主体——
// 左栏：8 组参数表单（垂直一排、整体滚动，「服务」组下含推理程序行）；
// 右栏：上 30% 状态区（状态行 + 校验行）、下 70% 日志区（VSplit 可拖动）。
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

// logMinHeight 日志区最小高度。
const logMinHeight = 200

// buildLayout 组装窗口内容并设置到 window（主线程，NewApp 构造期调用）。
func (a *App) buildLayout(window fyne.Window) {
	logArea := a.logv.Content() // *container.Scroll，可 SetMinSize
	logArea.SetMinSize(fyne.NewSize(0, logMinHeight))
	toolbar := container.NewHBox(a.form.saveBtn, a.form.loadBtn, a.status.Buttons(), layout.NewSpacer(), a.msgLbl)
	left := container.NewBorder(nil, nil, nil, nil, a.form.Content())
	right := container.NewVSplit(
		container.NewVBox(a.status.StatusRow(), a.form.errLabel),
		logArea)
	right.Offset = 0.3 // 状态区占右栏 30%，日志占 70%，分割条可拖动
	split := container.NewHSplit(left, right)
	split.Offset = 0.5 // 左右栏各占一半，分割条可拖动调整
	window.SetContent(container.NewBorder(toolbar, nil, nil, nil, split))
}
