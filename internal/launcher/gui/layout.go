// layout.go 组装窗口布局：顶部（校验行 + 操作行）/ 两栏主体——
// 左栏：选择 ninfer.exe 按钮 + 8 组参数 tab（字段垂直排列）；
// 右栏：上 30% 状态行、下 70% 日志区（VSplit 可拖动）。
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
	toolbar := container.NewHBox(a.form.saveBtn, a.form.loadBtn, layout.NewSpacer(), a.msgLbl)
	top := container.NewVBox(a.form.errLabel, toolbar)
	left := container.NewBorder(a.pickerBtn, nil, nil, nil, a.form.tabs)
	right := container.NewVSplit(a.status.Row(), logArea)
	right.Offset = 0.3 // 状态行占右栏 30%，日志占 70%，分割条可拖动
	split := container.NewHSplit(left, right)
	split.Offset = 0.62 // 左栏表单占多数，分割条可拖动调整
	window.SetContent(container.NewBorder(top, nil, nil, nil, split))
}
