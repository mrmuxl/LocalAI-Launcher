// status.go 承载状态行（状态/PID/运行时长/最新日志行）与启动、停止按钮（D4 状态机映射）。
//
// 线程模型：所有方法均在主线程调用（由 App 经 fyne.Do 调度）。
package gui

import (
	"fmt"
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"llmlauncher/internal/process"
	"llmlauncher/internal/version"
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
	versionLabel *widget.Label
	repoLink     *widget.Hyperlink
	emailLabel   *widget.Label
	startBtn     *widget.Button
	stopBtn      *widget.Button
}

// NewStatusPanel 构建状态面板（初始为已停止）。
// onStart/onStop 绑定到按钮（主线程回调，立即返回）。
func NewStatusPanel(onStart, onStop func()) *StatusPanel {
	p := &StatusPanel{
		stateLabel:   widget.NewLabel("已停止"),
		pidLabel:     widget.NewLabel("-"),
		uptimeLabel:  widget.NewLabel("-"),
		lastLogLabel: widget.NewLabel("（暂无）"),
		versionLabel: widget.NewLabel("v" + version.Version),
		repoLink:     widget.NewHyperlink("mrmuxl/LocalAI-Launcher", parseRepoURL()),
		emailLabel:   widget.NewLabel(version.Email),
	}
	p.startBtn = widget.NewButton("启动", onStart)
	p.stopBtn = widget.NewButton("停止", onStop)
	p.stopBtn.Importance = widget.DangerImportance
	p.ApplyState(process.Stopped, 0, time.Time{}, "")
	return p
}

// parseRepoURL 解析项目主页 URL（包级常量，失败返回 nil 不致命）。
func parseRepoURL() *url.URL {
	u, err := url.Parse(version.RepoURL)
	if err != nil {
		return nil
	}
	return u
}

// statusLabelWidth 状态表格标签列统一宽度（状态/PID/运行时长/最新日志对齐）。
const statusLabelWidth = 76

// StatusRow 返回状态表格（grid 左右两栏：左 = 运行状态四项，右 = 版本/项目/邮箱）。
func (p *StatusPanel) StatusRow() fyne.CanvasObject {
	labelRow := func(l *widget.Label, v fyne.CanvasObject) fyne.CanvasObject {
		return container.NewHBox(newFixedWidth(l, statusLabelWidth), v)
	}
	left := container.NewVBox(
		labelRow(widget.NewLabel("状态"), p.stateLabel),
		labelRow(widget.NewLabel("PID"), p.pidLabel),
		labelRow(widget.NewLabel("运行时长"), p.uptimeLabel),
		labelRow(widget.NewLabel("最新日志"), p.lastLogLabel),
	)
	right := container.NewVBox(
		labelRow(widget.NewLabel("版本"), p.versionLabel),
		labelRow(widget.NewLabel("项目"), p.repoLink),
		labelRow(widget.NewLabel("邮箱"), p.emailLabel),
	)
	return container.NewGridWithColumns(2, left, right)
}

// Buttons 返回启动/停止按钮行（供右栏操作行使用）。
func (p *StatusPanel) Buttons() fyne.CanvasObject {
	return container.NewHBox(p.startBtn, p.stopBtn)
}

// ApplyState 按 D4 状态机刷新文案并切换按钮可用性（主线程）。
// startedAt 为零值时显示运行时长「-」。
func (p *StatusPanel) ApplyState(st process.State, pid int, startedAt time.Time, lastLog string) {
	p.stateLabel.SetText(stateZh[st])
	if pid > 0 {
		p.pidLabel.SetText(fmt.Sprintf("%d", pid))
	} else {
		p.pidLabel.SetText("-")
	}
	if st == process.Running && !startedAt.IsZero() {
		p.uptimeLabel.SetText(time.Since(startedAt).Truncate(time.Second).String())
	} else {
		p.uptimeLabel.SetText("-")
	}
	if lastLog != "" {
		p.lastLogLabel.SetText(clipRunes(lastLog, 48))
	} else {
		p.lastLogLabel.SetText("（暂无）")
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
