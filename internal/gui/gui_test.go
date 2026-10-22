// gui_test.go 是 gui 包冒烟测试（T6 验收）：使用 fyne test 包在 headless 驱动下
// 构造窗口/表单/状态面板，验证：
//  1. 窗口构造不 panic；
//  2. 表单控件数 = 33（9 组）；
//  3. 状态机 D4 过渡态下启动按钮禁用。
package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"llmlauncher/internal/process"
)

// newTestWindow 初始化 headless 测试应用并返回一个空测试窗口。
func newTestWindow(t *testing.T) fyne.Window {
	t.Helper()
	test.NewApp()
	return test.NewWindow(nil)
}

// buildWindowContent 构造部分填充的 App 并调用真实布局方法 buildLayout，
// 保证测试与 app.go 布局零漂移。
func buildWindowContent(t *testing.T) fyne.Window {
	t.Helper()
	w := newTestWindow(t)

	form := NewNinferForm(w)
	llamaForm := NewLlamaForm(w)
	ring := process.NewRingLog(0)
	logv := NewLogView(ring)
	msgLbl := widget.NewLabel("")
	msgLbl.Wrapping = fyne.TextWrapWord
	status := NewStatusPanel(func() {}, func() {})

	a := &App{form: form, llamaForm: llamaForm, status: status, logv: logv, msgLbl: msgLbl}
	a.buildLayout(w)
	return w
}

// TestWindowConstruction 窗口构造（表单 + 状态 + 日志 + 布局）不得 panic。
func TestWindowConstruction(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("窗口构造 panic: %v", r)
		}
	}()
	w := buildWindowContent(t)
	if w == nil {
		t.Fatal("窗口构造返回 nil")
	}
}

// TestFormFieldCount 表单控件数必须为 33（9 组）。
func TestFormFieldCount(t *testing.T) {
	w := newTestWindow(t)
	f := NewNinferForm(w)

	if got := f.CountFields(); got != 33 {
		t.Fatalf("字段定义数 = %d，期望 33", got)
	}
	if got := len(f.entries) + len(f.checks) + len(f.selects); got != 33 {
		t.Fatalf("实际控件数 = %d（Entry %d + Check %d + Select %d），期望 33",
			got, len(f.entries), len(f.checks), len(f.selects))
	}
	if got := len(f.GroupNames()); got != 9 {
		t.Fatalf("分组数 = %d，期望 9", got)
	}
}

// TestLlamaFormFieldCount llama 表单控件数必须为 49（9 组）。
func TestLlamaFormFieldCount(t *testing.T) {
	w := newTestWindow(t)
	f := NewLlamaForm(w)

	if got := f.CountFields(); got != 49 {
		t.Fatalf("llama 字段定义数 = %d，期望 49", got)
	}
	if got := len(f.entries) + len(f.checks) + len(f.selects); got != 49 {
		t.Fatalf("llama 实际控件数 = %d（Entry %d + Check %d + Select %d），期望 49",
			got, len(f.entries), len(f.checks), len(f.selects))
	}
	if got := len(f.GroupNames()); got != 9 {
		t.Fatalf("llama 分组数 = %d，期望 9", got)
	}
}

// TestStartButtonDisabledInTransition 状态机 D4：过渡态（starting/stopping）禁用启停按钮。
func TestStartButtonDisabledInTransition(t *testing.T) {
	w := newTestWindow(t)
	_ = w
	p := NewStatusPanel(func() {}, func() {})

	// 初始 stopped：启动可用、停止禁用。
	if p.StartDisabled() {
		t.Fatal("stopped: 启动按钮应可用")
	}
	if !p.StopDisabled() {
		t.Fatal("stopped: 停止按钮应禁用")
	}

	// starting：两按钮均禁用（T6 验收核心断言）。
	p.ApplyState(process.Starting, 0, time.Time{}, "")
	if !p.StartDisabled() {
		t.Fatal("starting: 启动按钮应禁用")
	}
	if !p.StopDisabled() {
		t.Fatal("starting: 停止按钮应禁用")
	}

	// running：启动禁用、停止可用。
	p.ApplyState(process.Running, 123, time.Now(), "log line")
	if !p.StartDisabled() {
		t.Fatal("running: 启动按钮应禁用")
	}
	if p.StopDisabled() {
		t.Fatal("running: 停止按钮应可用")
	}

	// stopping：两按钮均禁用。
	p.ApplyState(process.Stopping, 123, time.Now(), "log line")
	if !p.StartDisabled() || !p.StopDisabled() {
		t.Fatal("stopping: 两按钮均应禁用")
	}
}

// TestIsAtBottom 表驱动测试 sticky-bottom 判定纯函数。
func TestIsAtBottom(t *testing.T) {
	tests := []struct {
		name           string
		contentHeight  float32
		viewportHeight float32
		offset         float32
		want           bool
	}{
		{"尚未布局：视口高度为 0", 500, 0, 0, true},
		{"尚未布局：内容高度为 0", 0, 300, 0, true},
		{"内容不超出视口", 200, 300, 0, true},
		{"已在底部", 1000, 300, 700, true},
		{"距底部 39px（阈值内）", 1000, 300, 661, true},
		{"距底部 40px（恰好阈值）", 1000, 300, 660, true},
		{"距底部 41px（超出阈值）", 1000, 300, 659, false},
		{"在顶部", 1000, 300, 0, false},
		{"内容恰好等于视口+1px", 301, 300, 1, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAtBottom(tc.contentHeight, tc.viewportHeight, tc.offset); got != tc.want {
				t.Fatalf("isAtBottom(%v, %v, %v) = %v，期望 %v",
					tc.contentHeight, tc.viewportHeight, tc.offset, got, tc.want)
			}
		})
	}
}

// TestLogViewRefreshNoPanic 验证 Refresh 在有日志内容时不 panic 且文本正确更新。
func TestLogViewRefreshNoPanic(t *testing.T) {
	t.Helper()
	test.NewApp()

	ring := process.NewRingLog(10)
	ring.Append("line 1")
	ring.Append("line 2")
	ring.Append("line 3")

	lv := NewLogView(ring)
	scroll := lv.Content()
	if scroll == nil {
		t.Fatal("Content() 返回 nil")
	}

	// 首次刷新：文本应从占位符变为实际日志
	lv.Refresh()
	want := "line 1\nline 2\nline 3"
	if lv.seg.Text != want {
		t.Fatalf("Refresh 后 seg.Text = %q，期望 %q", lv.seg.Text, want)
	}

	// 追加新行后再次刷新
	ring.Append("line 4")
	lv.Refresh()
	want = "line 1\nline 2\nline 3\nline 4"
	if lv.seg.Text != want {
		t.Fatalf("追加后 seg.Text = %q，期望 %q", lv.seg.Text, want)
	}

	// 空 ring 回退到占位符
	emptyRing := process.NewRingLog(10)
	lv2 := NewLogView(emptyRing)
	lv2.Content()
	lv2.Refresh()
	if lv2.seg.Text != "（暂无日志输出）" {
		t.Fatalf("空 ring 刷新后 seg.Text = %q，期望占位符", lv2.seg.Text)
	}
}

// TestLogViewSetRingForceBottom 验证 SetRing 切换数据源后正常刷新不 panic。
func TestLogViewSetRingForceBottom(t *testing.T) {
	t.Helper()
	test.NewApp()

	ringA := process.NewRingLog(10)
	ringA.Append("A line 1")
	ringA.Append("A line 2")

	ringB := process.NewRingLog(10)
	ringB.Append("B line 1")

	lv := NewLogView(ringA)
	lv.Content()

	lv.SetRing(ringB)
	if lv.seg.Text != "B line 1" {
		t.Fatalf("SetRing 后 seg.Text = %q，期望 %q", lv.seg.Text, "B line 1")
	}
}

// TestTabsCountAndTitle 验证左栏 AppTabs 共 3 项且第三项标题为「说明」。
func TestTabsCountAndTitle(t *testing.T) {
	w := newTestWindow(t)
	form := NewNinferForm(w)
	llamaForm := NewLlamaForm(w)
	ring := process.NewRingLog(0)
	logv := NewLogView(ring)
	msgLbl := widget.NewLabel("")
	msgLbl.Wrapping = fyne.TextWrapWord
	status := NewStatusPanel(func() {}, func() {})

	a := &App{form: form, llamaForm: llamaForm, status: status, logv: logv, msgLbl: msgLbl}
	a.buildLayout(w)

	if a.tabs == nil {
		t.Fatal("tabs 为 nil")
	}
	items := a.tabs.Items
	if len(items) != 3 {
		t.Fatalf("tabs 项数 = %d，期望 3", len(items))
	}
	if items[2].Text != "说明" {
		t.Fatalf("第三项标题 = %q，期望「说明」", items[2].Text)
	}
}
