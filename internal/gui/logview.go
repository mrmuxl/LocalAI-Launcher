// logview.go 把进程日志环形缓冲区（最近 500 行）渲染为只读滚动文本区。
//
// 线程模型：后台 goroutine 只向 RingLog 追加（其自身线程安全）；
// 主线程定时器调用 Refresh 更新控件，goroutine 内不触碰 Fyne 控件。
//
// Sticky-Bottom 行为：
//   - Refresh 前先判断视口是否贴底（距底部 ≤ 2 行高度）；
//   - 贴底 → 更新文本后调用 ScrollToBottom 始终跟随最新日志；
//   - 不贴底（用户正在翻历史）→ 保持当前滚动位置，不打断阅读；
//   - SetRing 切 tab 时强制回到底部。
package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"llmlauncher/internal/process"
)

// stickyBottomThreshold 贴底判定阈值（像素）：视口距内容底部不超过此值视为贴底。
// 取 40px ≈ 2 行文本高度（默认字号下每行约 20px），兼顾视觉连续性。
const stickyBottomThreshold float32 = 40

// LogView 是只读日志视图。
type LogView struct {
	rich   *widget.RichText
	seg    *widget.TextSegment
	ring   *process.RingLog
	scroll *container.Scroll
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
// 保存 Scroll 引用以便 Refresh 时读取/设置滚动位置。
func (lv *LogView) Content() *container.Scroll {
	s := container.NewScroll(lv.rich)
	s.SetMinSize(fyne.NewSize(0, 180))
	lv.scroll = s
	return s
}

// SetRing 切换数据源并立即重绘（主线程调用）：tab 切换时换到对应引擎的 ring。
// 切 tab 后强制回到底部，确保用户看到最新日志。
func (lv *LogView) SetRing(r *process.RingLog) {
	lv.ring = r
	lv.Refresh()
	if lv.scroll != nil {
		lv.scroll.ScrollToBottom()
	}
}

// Refresh 渲染环形缓冲区内容（主线程调用；内容未变化时跳过刷新）。
// Sticky-Bottom：刷新前记录视口是否贴底，更新文本后若贴底则自动滚动到底部。
func (lv *LogView) Refresh() {
	lines := lv.ring.Snapshot()
	text := strings.Join(lines, "\n")
	if text == "" {
		text = "（暂无日志输出）"
	}
	if text == lv.seg.Text {
		return
	}

	// 刷新前检测视口是否贴底
	atBottom := true // 默认贴底（scroll 尚未构建时兜底）
	if lv.scroll != nil {
		atBottom = isAtBottom(
			lv.rich.MinSize().Height,
			lv.scroll.Size().Height,
			lv.scroll.Offset.Y,
		)
	}

	lv.seg.Text = text
	lv.rich.Refresh()

	// 若刷新前视口贴底，更新后自动跟随到底部
	if atBottom && lv.scroll != nil {
		lv.scroll.ScrollToBottom()
	}
}

// isAtBottom 判断视口是否贴底：内容高度 - 当前偏移 - 视口高度 ≤ 阈值时视为贴底。
// 尚未布局（视口高度 ≤ 0 或内容高度 ≤ 0）时视为贴底，避免首次刷新丢失跟随。
func isAtBottom(contentHeight, viewportHeight, offset float32) bool {
	if viewportHeight <= 0 || contentHeight <= 0 {
		return true
	}
	remaining := contentHeight - offset - viewportHeight
	return remaining <= stickyBottomThreshold
}
