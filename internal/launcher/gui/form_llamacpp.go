// form_llamacpp.go llama.cpp 引擎的表单（18 字段 / 7 组）。
//
// 与 ninfer 表单（form_ninfer.go）完全分离：字段集、默认值、argv flag
// 都不同（对应 start-backend.ps1 的 llamacpp 分支）。复用同包的
// newFixedWidth/entryWidth/collectInt/collectFloat 布局与解析工具。
package gui

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"ninferrmanager/internal/launcher/config"
)

// llamaFieldDef 是 llama 表单对单个字段的唯一声明。
type llamaFieldDef struct {
	group string
	name  string
	label string
	kind  fieldKind
}

// llamaFieldDefs：18 字段，7 组，顺序与 config.LlamaConfig 字段顺序一致。
var llamaFieldDefs = []llamaFieldDef{
	// 服务（2）
	{"服务", "Host", "监听地址", kindString},
	{"服务", "Port", "监听端口", kindInt},
	// 模型（1）
	{"模型", "ModelPath", "模型文件（GGUF 路径，-m）", kindString},
	// 上下文与内存（4）
	{"上下文与内存", "Context", "上下文长度（-c）", kindInt},
	{"上下文与内存", "Budget", "KV 预算（--kvmem-budget）", kindInt},
	{"上下文与内存", "Reserve", "生成保留（--kvmem-gen-reserve）", kindInt},
	{"上下文与内存", "HostMiB", "Host KV 预算 MiB（0 = 不传）", kindInt},
	// KV 缓存（3）
	{"KV 缓存", "KvType", "KV 类型（--kv-dtype）", kindString},
	{"KV 缓存", "SinkTokens", "sink 保留 token", kindInt},
	{"KV 缓存", "RecentTokens", "recent 保留 token", kindInt},
	// 生成（2）
	{"生成", "MaxTokens", "最大生成 token（-n）", kindInt},
	{"生成", "Prefill", "prefill 批大小（-b/-ub）", kindInt},
	// LRFU（4）
	{"LRFU", "NoLrfu", "关闭 LRFU 历史加权", kindBool},
	{"LRFU", "LrfuLambda", "LRFU lambda [0,1]", kindFloat},
	{"LRFU", "LrfuTopFraction", "LRFU top-frac (0,1]", kindFloat},
	{"LRFU", "ScoreAggregation", "GQA 聚合（mean/max）", kindString},
	// SSD 分层（2）
	{"SSD 分层", "DiskPath", "SSD 目录（空 = 关闭）", kindString},
	{"SSD 分层", "DiskMiB", "SSD 预算 MiB", kindInt},
}

// LlamaForm 持有 llama.cpp 表单全部控件（18 个）与校验状态行。
// 所有方法必须运行在主线程（Fyne 约束）。
type LlamaForm struct {
	window   fyne.Window
	content  fyne.CanvasObject
	entries  map[string]*widget.Entry
	checks   map[string]*widget.Check
	errLabel *widget.Label

	workerPathEntry *widget.Entry // llama worker 可执行文件路径输入框

	onChange        func()            // App 注入：字段变化时触发后台校验
	onOpenWorker    func()            // App 注入：打开按钮 → 文件选择框
	onWorkerPathSet func(path string) // App 注入：输入框回车提交路径
}

// NewLlamaForm 构建 7 组表单（垂直一排、带分组标题、整体滚动）并返回。
// 行布局与 ninfer 表单一致：统一标签列宽 + 固定输入框宽度。
func NewLlamaForm(window fyne.Window) *LlamaForm {
	f := &LlamaForm{
		window:  window,
		entries: make(map[string]*widget.Entry, len(llamaFieldDefs)),
		checks:  make(map[string]*widget.Check),
	}

	controls := make(map[string]fyne.CanvasObject, len(llamaFieldDefs))
	for _, d := range llamaFieldDefs {
		switch d.kind {
		case kindBool:
			c := widget.NewCheck("", nil)
			c.OnChanged = func(bool) { f.notifyChange() }
			f.checks[d.name] = c
			controls[d.name] = c
		default:
			e := widget.NewEntry()
			e.OnChanged = func(string) { f.notifyChange() }
			f.entries[d.name] = e
			controls[d.name] = e
		}
	}

	// 统一标签列宽：取全部行标签（含「推理程序」）的最大文本宽度。
	labelWidth := widget.NewLabel("推理程序").MinSize().Width
	for _, d := range llamaFieldDefs {
		if w := widget.NewLabel(d.label).MinSize().Width; w > labelWidth {
			labelWidth = w
		}
	}

	row := func(label string, w fyne.CanvasObject, fixed bool) fyne.CanvasObject {
		if fixed {
			w = newFixedWidth(w, entryWidth)
		}
		return container.NewHBox(newFixedWidth(widget.NewLabel(label), labelWidth), w, layout.NewSpacer())
	}

	var sections []fyne.CanvasObject
	seen := make(map[string]bool, 7)
	for _, d := range llamaFieldDefs {
		if !seen[d.group] {
			seen[d.group] = true
			sections = append(sections,
				widget.NewLabelWithStyle(d.group, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
			if d.group == "服务" {
				sections = append(sections, f.buildWorkerPathRow(labelWidth))
			}
		}
		_, isCheck := f.checks[d.name]
		sections = append(sections, row(d.label, controls[d.name], !isCheck))
	}
	f.content = container.NewScroll(container.NewVBox(sections...))

	f.errLabel = widget.NewLabel("✓ 配置有效")
	f.errLabel.Wrapping = fyne.TextWrapWord
	return f
}

// Content 返回表单滚动容器（供窗口布局使用）。
func (f *LlamaForm) Content() fyne.CanvasObject {
	return f.content
}

// ErrLabel 返回校验状态行（供窗口布局使用）。
func (f *LlamaForm) ErrLabel() fyne.CanvasObject {
	return f.errLabel
}

// buildWorkerPathRow 构建「推理程序」行：标签 + 固定宽路径输入框 + 打开按钮。
func (f *LlamaForm) buildWorkerPathRow(labelWidth float32) fyne.CanvasObject {
	f.workerPathEntry = widget.NewEntry()
	f.workerPathEntry.SetPlaceHolder("llama worker exe 路径（输入后回车，或点「打开」选择）")
	f.workerPathEntry.OnSubmitted = func(text string) {
		if f.onWorkerPathSet != nil {
			f.onWorkerPathSet(text)
		}
	}
	openBtn := widget.NewButton("打开", func() {
		if f.onOpenWorker != nil {
			f.onOpenWorker()
		}
	})
	return container.NewHBox(
		newFixedWidth(widget.NewLabel("推理程序"), labelWidth),
		newFixedWidth(f.workerPathEntry, entryWidth),
		openBtn,
		layout.NewSpacer(),
	)
}

// SetWorkerPath 回填推理程序路径输入框（主线程；SetText 不触发 OnSubmitted）。
func (f *LlamaForm) SetWorkerPath(p string) {
	if f.workerPathEntry != nil {
		f.workerPathEntry.SetText(p)
	}
}

// notifyChange 把字段变化转发给 App（主线程内立即返回）。
func (f *LlamaForm) notifyChange() {
	if f.onChange != nil {
		f.onChange()
	}
}

// CountFields 返回字段总数（应为 18）。
func (f *LlamaForm) CountFields() int { return len(llamaFieldDefs) }

// SetFieldsFrom 用配置填充全部控件（主线程）。
func (f *LlamaForm) SetFieldsFrom(cfg config.LlamaConfig) {
	f.entries["Host"].SetText(cfg.Host)
	f.entries["Port"].SetText(strconv.Itoa(cfg.Port))
	f.entries["ModelPath"].SetText(cfg.ModelPath)
	f.entries["Context"].SetText(strconv.Itoa(cfg.Context))
	f.entries["Budget"].SetText(strconv.Itoa(cfg.Budget))
	f.entries["Reserve"].SetText(strconv.Itoa(cfg.Reserve))
	f.entries["HostMiB"].SetText(strconv.Itoa(cfg.HostMiB))
	f.entries["KvType"].SetText(cfg.KvType)
	f.entries["SinkTokens"].SetText(strconv.Itoa(cfg.SinkTokens))
	f.entries["RecentTokens"].SetText(strconv.Itoa(cfg.RecentTokens))
	f.entries["MaxTokens"].SetText(strconv.Itoa(cfg.MaxTokens))
	f.entries["Prefill"].SetText(strconv.Itoa(cfg.Prefill))
	f.checks["NoLrfu"].SetChecked(cfg.NoLrfu)
	f.entries["LrfuLambda"].SetText(strconv.FormatFloat(cfg.LrfuLambda, 'f', -1, 64))
	f.entries["LrfuTopFraction"].SetText(strconv.FormatFloat(cfg.LrfuTopFraction, 'f', -1, 64))
	f.entries["ScoreAggregation"].SetText(cfg.ScoreAggregation)
	f.entries["DiskPath"].SetText(cfg.DiskPath)
	f.entries["DiskMiB"].SetText(strconv.Itoa(cfg.DiskMiB))
}

// CollectConfig 从控件收集当前配置（主线程轻量读）。
// 第二个返回值是中文解析错误列表（空串按默认值处理）。
func (f *LlamaForm) CollectConfig() (config.LlamaConfig, []string) {
	var cfg config.LlamaConfig
	var errs []string

	cfg.Host = f.entries["Host"].Text
	cfg.Port, errs = collectInt(f.entries["Port"], "监听端口", cfg.Port, errs)
	cfg.ModelPath = f.entries["ModelPath"].Text
	cfg.Context, errs = collectInt(f.entries["Context"], "上下文长度", cfg.Context, errs)
	cfg.Budget, errs = collectInt(f.entries["Budget"], "KV 预算", cfg.Budget, errs)
	cfg.Reserve, errs = collectInt(f.entries["Reserve"], "生成保留", cfg.Reserve, errs)
	cfg.HostMiB, errs = collectInt(f.entries["HostMiB"], "Host KV 预算", cfg.HostMiB, errs)
	cfg.KvType = f.entries["KvType"].Text
	cfg.SinkTokens, errs = collectInt(f.entries["SinkTokens"], "sink 保留", cfg.SinkTokens, errs)
	cfg.RecentTokens, errs = collectInt(f.entries["RecentTokens"], "recent 保留", cfg.RecentTokens, errs)
	cfg.MaxTokens, errs = collectInt(f.entries["MaxTokens"], "最大生成 token", cfg.MaxTokens, errs)
	cfg.Prefill, errs = collectInt(f.entries["Prefill"], "prefill 批大小", cfg.Prefill, errs)
	cfg.NoLrfu = f.checks["NoLrfu"].Checked
	cfg.LrfuLambda, errs = collectFloat(f.entries["LrfuLambda"], "LRFU lambda", cfg.LrfuLambda, errs)
	cfg.LrfuTopFraction, errs = collectFloat(f.entries["LrfuTopFraction"], "LRFU top-frac", cfg.LrfuTopFraction, errs)
	cfg.ScoreAggregation = f.entries["ScoreAggregation"].Text
	cfg.DiskPath = f.entries["DiskPath"].Text
	cfg.DiskMiB, errs = collectInt(f.entries["DiskMiB"], "SSD 预算", cfg.DiskMiB, errs)
	return cfg, errs
}

// SetValidationErrors 显示校验结果（中文，每行一条；主线程）。
func (f *LlamaForm) SetValidationErrors(errs []string) {
	if len(errs) == 0 {
		f.errLabel.SetText("✓ 配置有效")
		return
	}
	var b strings.Builder
	for _, e := range errs {
		b.WriteString("✗ ")
		b.WriteString(e)
		b.WriteByte('\n')
	}
	f.errLabel.SetText(strings.TrimSuffix(b.String(), "\n"))
}
