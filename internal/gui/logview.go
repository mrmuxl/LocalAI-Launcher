// logview.go 把进程日志环形缓冲区（最近 500 行）渲染为只读滚动文本区。
//
// 线程模型：后台 goroutine 只向 RingLog 追加（其自身线程安全）；
// 主线程定时器调用 Refresh 更新控件，goroutine 内不触碰 Fyne 控件。
package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"llmlauncher/internal/process"
)

// LogView 是只读日志视图。
type LogView struct {
	rich *widget.RichText
	seg  *widget.TextSegment
	ring *process.RingLog
}

// NewLogView 构建日志视图；ring 与进程管理器共享同一实例。
// RichText 开启按词自动换行：长日志行在容器宽度内折行，不横向溢出。
func NewLogView(ring *process.RingLog) *LogView {
	seg := &widget.TextSegment{Text: "（暂无日志输出）"}
	rich := widget.NewRichText(seg)
	rich.Wrapping = fyne.TextWrapWord
	return &LogView{
		rich: rich,
		seg:  seg,
		ring: ring,
	}
}

// Content 是日志滚动区（供窗口布局使用）。
func (lv *LogView) Content() *container.Scroll {
	s := container.NewScroll(lv.rich)
	s.SetMinSize(fyne.NewSize(0, 180))
	return s
}

// Refresh 渲染环形缓冲区内容（主线程调用；内容未变化时跳过刷新）。
func (lv *LogView) Refresh() {
	lines := lv.ring.Snapshot()
	text := strings.Join(lines, "\n")
	if text == "" {
		text = "（暂无日志输出）"
	}
	if text != lv.seg.Text {
		lv.seg.Text = text
		lv.rich.Refresh()
	}
}
