// status.go 承载状态行（状态/PID/运行时长/最新日志行）与启动、停止按钮（D4 状态机映射）。
//
// 线程模型：所有方法均在主线程调用（由 App 经 fyne.Do 调度）。
package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"ninferrmanager/internal/launcher/process"
)

// stateZh 状态枚举 → 中文文案（状态机 D4）。
var stateZh = map[process.State]string{
	process.Stopped:  "已停止",
	process.Starting: "启动中…",
	process.Running:  "运行中",
	process.Stopping: "停止中…",
}

// StatusPanel 是状态行 + 启停按钮的组合部件。
type StatusPanel struct {
	stateLabel   *widget.Label
	pidLabel     *widget.Label
	uptimeLabel  *widget.Label
	lastLogLabel *widget.Label
	startBtn     *widget.Button
	stopBtn      *widget.Button
}

// NewStatusPanel 构建状态面板（初始为已停止）。
// onStart/onStop 绑定到按钮（主线程回调，立即返回）。
func NewStatusPanel(onStart, onStop func()) *StatusPanel {
	p := &StatusPanel{
		stateLabel:   widget.NewLabel("状态：已停止"),
		pidLabel:     widget.NewLabel("PID：-"),
		uptimeLabel:  widget.NewLabel("运行时长：-"),
		lastLogLabel: widget.NewLabel("最新日志：（暂无）"),
	}
	p.startBtn = widget.NewButton("启动", onStart)
	p.stopBtn = widget.NewButton("停止", onStop)
	p.stopBtn.Importance = widget.DangerImportance
	p.ApplyState(process.Stopped, 0, time.Time{}, "")
	return p
}

// Row 返回状态行（供窗口布局使用）。
func (p *StatusPanel) Row() fyne.CanvasObject {
	return container.NewHBox(
		p.stateLabel,
		p.pidLabel,
		p.uptimeLabel,
		p.lastLogLabel,
		layout.NewSpacer(),
		p.startBtn,
		p.stopBtn,
	)
}

// ApplyState 按 D4 状态机刷新文案并切换按钮可用性（主线程）。
// startedAt 为零值时显示运行时长「-」。
func (p *StatusPanel) ApplyState(st process.State, pid int, startedAt time.Time, lastLog string) {
	p.stateLabel.SetText("状态：" + stateZh[st])
	if pid > 0 {
		p.pidLabel.SetText(fmt.Sprintf("PID：%d", pid))
	} else {
		p.pidLabel.SetText("PID：-")
	}
	if st == process.Running && !startedAt.IsZero() {
		p.uptimeLabel.SetText("运行时长：" + time.Since(startedAt).Truncate(time.Second).String())
	} else {
		p.uptimeLabel.SetText("运行时长：-")
	}
	if lastLog != "" {
		p.lastLogLabel.SetText("最新日志：" + clipRunes(lastLog, 48))
	} else {
		p.lastLogLabel.SetText("最新日志：（暂无）")
	}

	switch st {
	case process.Stopped:
		p.startBtn.Enable()
		p.stopBtn.Disable()
	case process.Starting:
		p.startBtn.Disable()
		p.stopBtn.Disable()
	case process.Running:
		p.startBtn.Disable()
		p.stopBtn.Enable()
	case process.Stopping:
		p.startBtn.Disable()
		p.stopBtn.Disable()
	}
}

// StartDisabled/StopDisabled 供测试断言按钮初态。
func (p *StatusPanel) StartDisabled() bool { return p.startBtn.Disabled() }
func (p *StatusPanel) StopDisabled() bool  { return p.stopBtn.Disabled() }

// clipRunes 截断字符串到至多 n 个 rune（避免截断多字节字符）。
func clipRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
