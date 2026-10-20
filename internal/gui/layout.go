// layout.go 组装窗口布局：两栏主体——
// 左栏：引擎切换 AppTabs（NInfer / llama.cpp，各自表单整体滚动，
// 「服务」组下含推理引擎行）；右栏：上 30% 状态区（操作行 + 操作反馈 +
// 状态表格 + 当前引擎校验行）、下 70% 日志区（VSplit 可拖动）。
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// logMinHeight 日志区最小高度。
const logMinHeight = 200

// buildLayout 组装窗口内容并设置到 window（主线程，NewApp 构造期调用）。
func (a *App) buildLayout(window fyne.Window) {
	logArea := a.logv.Content() // *container.Scroll，可 SetMinSize
	logArea.SetMinSize(fyne.NewSize(0, logMinHeight))

	// 左栏双引擎 tab：校验行跟随当前选中 tab 显示。
	tabs := container.NewAppTabs(
		container.NewTabItem("NInfer", a.form.Content()),
		container.NewTabItem("llama.cpp", a.llamaForm.Content()),
	)
	a.tabs = tabs
	errBox := container.NewVBox(a.form.errLabel, a.llamaForm.errLabel)
	syncErrVisible := func() {
		a.form.errLabel.Hidden = tabs.SelectedIndex() != 0
		a.llamaForm.errLabel.Hidden = tabs.SelectedIndex() != 1
		errBox.Refresh()
	}
	a.form.errLabel.Hidden = false
	a.llamaForm.errLabel.Hidden = true
	tabs.OnSelected = func(*container.TabItem) { syncErrVisible() }
	syncErrVisible()

	// 右栏顶部操作行：保存配置/重新加载/启动/停止（onSave/onLoad 已按当前 tab 分流）；
	// 操作反馈（保存/加载/路径保存等消息）紧随操作行显示在右栏。
	actions := container.NewHBox(
		widget.NewButton("保存配置", a.onSave),
		widget.NewButton("重新加载", a.onLoad),
		a.status.Buttons(),
		layout.NewSpacer(),
	)

	left := container.NewBorder(nil, nil, nil, nil, tabs)
	right := container.NewVSplit(
		container.NewVBox(actions, a.msgLbl, a.status.StatusRow(), errBox),
		logArea)
	right.Offset = 0.3 // 状态区占右栏 30%，日志占 70%，分割条可拖动
	split := container.NewHSplit(left, right)
	split.Offset = 0.5 // 左右栏各占一半，分割条可拖动调整
	window.SetContent(container.NewBorder(nil, nil, nil, nil, split))
}
