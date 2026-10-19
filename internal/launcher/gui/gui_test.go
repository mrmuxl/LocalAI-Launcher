// gui_test.go 是 gui 包冒烟测试（T6 验收）：使用 fyne test 包在 headless 驱动下
// 构造窗口/表单/状态面板，验证：
//  1. 窗口构造不 panic；
//  2. 表单控件数 = 34（8 组）；
//  3. 状态机 D4 过渡态下启动按钮禁用。
package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"ninferrmanager/internal/launcher/process"
)

// newTestWindow 初始化 headless 测试应用并返回一个空测试窗口。
func newTestWindow(t *testing.T) fyne.Window {
	t.Helper()
	test.NewApp()
	return test.NewWindow(nil)
}

// buildWindowContent 复刻 NewApp 的窗口内容装配（app.go 布局段），
// 供「窗口构造不 panic」断言使用。
func buildWindowContent(t *testing.T) fyne.Window {
	t.Helper()
	w := newTestWindow(t)

	form := NewForm(w)
	ring := process.NewRingLog(0)
	logv := NewLogView(ring)
	msgLbl := widget.NewLabel("")
	msgLbl.Wrapping = fyne.TextWrapWord
	status := NewStatusPanel(func() {}, func() {})

	// 与 NewApp 相同的布局：顶部（校验行 + 操作行）/ 两栏主体——
	// 左栏：选择按钮 + 参数 tab；右栏：上 30% 状态行、下 70% 日志区。
	logArea := logv.Content()
	logArea.SetMinSize(fyne.NewSize(0, logMinHeight))
	pickerBtn := widget.NewButton("选择 ninfer.exe…", func() {})
	toolbar := container.NewHBox(widget.NewButton("保存", func() {}), widget.NewButton("重新加载", func() {}), layout.NewSpacer(), msgLbl)
	top := container.NewVBox(form.errLabel, toolbar)
	left := container.NewBorder(pickerBtn, nil, nil, nil, form.tabs)
	right := container.NewVSplit(status.Row(), logArea)
	right.Offset = 0.3
	split := container.NewHSplit(left, right)
	split.Offset = 0.62
	w.SetContent(container.NewBorder(top, nil, nil, nil, split))
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

// TestFormFieldCount 表单控件数必须为 34（30 Entry + 4 Check，8 组）。
func TestFormFieldCount(t *testing.T) {
	w := newTestWindow(t)
	f := NewForm(w)

	if got := f.CountFields(); got != 34 {
		t.Fatalf("字段定义数 = %d，期望 34", got)
	}
	if got := len(f.entries) + len(f.checks); got != 34 {
		t.Fatalf("实际控件数 = %d（Entry %d + Check %d），期望 34", got, len(f.entries), len(f.checks))
	}
	if got := len(f.GroupNames()); got != 8 {
		t.Fatalf("分组数 = %d，期望 8", got)
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
