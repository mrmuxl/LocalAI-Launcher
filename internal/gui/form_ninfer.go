// form_ninfer.go 构建 32 字段 / 9 分组的 NInfer 启动器表单，并承载校验交互。
//
// 字段定义与 config.Config 的 34 个字段一一对应（名称/顺序不可增减）；
// 中文标签取自 config.go 字段注释（唯一来源）。
package gui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"llmlauncher/internal/config"
)

// fieldKind 描述字段输入类型。
type fieldKind int

const (
	kindString fieldKind = iota
	kindInt
	kindFloat
	kindBool
	kindSelect
)

// fieldDef 是表单对单个字段的唯一声明。
type fieldDef struct {
	group   string
	name    string
	label   string
	kind    fieldKind
	options []string // kindSelect 的选项
}

// fieldDefs：32 字段，9 组，顺序与 config.Config 字段顺序一致。
// 字段与默认值对齐 start-ninfer.bat 默认配方（真实 ninfer-kvmem-server.exe 参数）；
// 标签末尾括号内为对应命令行参数（来源：docs/命令行参数参考.txt 与 argv 构造契约）。
var fieldDefs = []fieldDef{
	// 服务（3）
	{"服务", "ListenHost", "监听地址（--host）", kindString, nil},
	{"服务", "Port", "监听端口（--port）", kindInt, nil},
	{"服务", "Concurrency", "并发解码请求数（--max-concurrency，1..8）", kindInt, nil},
	// 模型（3）
	{"模型", "ModelPath", "模型文件（位置参数 model.ninfer）", kindString, nil},
	{"模型", "Gpu", "GPU 编号/UUID（CUDA_VISIBLE_DEVICES，逗号分隔）", kindString, nil},
	{"模型", "StageLayers", "每卡层数（--stage-layers，空 = 单卡）", kindString, nil},
	// 上下文与内存（6）
	{"上下文与内存", "Context", "上下文上限（--max-context）", kindInt, nil},
	{"上下文与内存", "Budget", "历史预算 B（--kvmem-budget，64 对齐）", kindInt, nil},
	{"上下文与内存", "Reserve", "生成保留 R（--kvmem-gen-reserve，64 对齐）", kindInt, nil},
	{"上下文与内存", "HostMiB", "Host KV 预算（--kvmem-host-mib，MiB）", kindInt, nil},
	{"上下文与内存", "MaxTokens", "默认输出上限（--default-max-tokens）", kindInt, nil},
	{"上下文与内存", "Prefill", "prefill 块（--prefill-chunk，128 对齐）", kindInt, nil},
	// KV 缓存（5）
	{"KV 缓存", "KvType", "KV 存储（--kv-dtype）", kindSelect,
		[]string{"bf16", "int8", "fp8", "nvfp4", "k8v4", "rk8v4", "rk4v4", "rk4v4-e8", "rk2v4-e8"}},
	{"KV 缓存", "SinkTokens", "常驻前缀（--kvmem-sink-tokens）", kindInt, nil},
	{"KV 缓存", "RecentTokens", "常驻后缀（--kvmem-recent-tokens，≥64）", kindInt, nil},
	{"KV 缓存", "RetainedSessions", "保留历史数（--kvmem-sessions，1..16）", kindInt, nil},
	{"KV 缓存", "FastPrefill", "快速 prefill 内核（--fast-prefill-kernel）", kindBool, nil},
	// 推测解码（4）
	{"推测解码", "MtpDrafts", "MTP 草稿数（--draft-tokens，0 = 关闭）", kindInt, nil},
	{"推测解码", "NgramDrafts", "ngram 草稿数（--ngram-draft-tokens，0..63）", kindInt, nil},
	{"推测解码", "NgramMinMatch", "ngram 最短匹配（--ngram-min-match，4..64）", kindInt, nil},
	{"推测解码", "AdaptiveMtp", "自适应 MTP（--adaptive-mtp）", kindBool, nil},
	// 视觉（3）
	{"视觉", "Vision", "启用视觉（--vision）", kindBool, nil},
	{"视觉", "VisionResidency", "视觉驻留（--vision-residency）", kindSelect,
		[]string{"resident", "cpu"}},
	{"视觉", "VisionTokens", "单媒体合并上限（--vision-max-merged）", kindInt, nil},
	// 思考与设备（3）
	{"思考与设备", "ThinkingBudget", "默认思考预算（--default-thinking-budget，0 = 不传）", kindInt, nil},
	{"思考与设备", "DefaultReasoningEffort", "默认思考力度（--default-reasoning-effort，空 = 不传）", kindSelect,
		[]string{"none", "minimal", "low", "medium", "high", "xhigh", "max"}},
	{"思考与设备", "DeviceProfile", "设备档案（--device-profile）", kindSelect,
		[]string{"off", "auto"}},
	// SSD 分层（2）
	{"SSD 分层", "DiskPath", "SSD 分层目录（--kvmem-disk-path，空 = 不传）", kindString, nil},
	{"SSD 分层", "DiskMiB", "SSD 配额（--kvmem-disk-mib，MiB）", kindInt, nil},
	// 检索调优（4）
	{"检索调优", "NoLrfu", "关闭 LRFU（不勾选 = 启用 --kvmem-lrfu）", kindBool, nil},
	{"检索调优", "LrfuLambda", "历史衰减（--kvmem-lrfu-lambda，0..1）", kindFloat, nil},
	{"检索调优", "LrfuTopFraction", "命中比例（--kvmem-lrfu-top-frac，(0,1]）", kindFloat, nil},
	{"检索调优", "ScoreAggregation", "评分聚合（--kvmem-score-agg）", kindSelect,
		[]string{"mean", "max"}},
}

// NinferForm 持有全部表单控件（32 个）与校验状态行。
// 所有方法必须运行在主线程（Fyne 约束）。
type NinferForm struct {
	window   fyne.Window
	content  fyne.CanvasObject // 9 组表单垂直排列的滚动容器
	entries  map[string]*widget.Entry
	checks   map[string]*widget.Check
	selects  map[string]*widget.Select
	errLabel *widget.Label

	workerPathEntry *widget.Entry // 推理引擎路径输入框（服务组下方）

	onChange        func()            // App 注入：字段变化时触发后台校验
	onOpenWorker    func()            // App 注入：打开按钮 → 文件选择框
	onWorkerPathSet func(path string) // App 注入：输入框回车提交路径
	onOpenModel     func()            // App 注入：模型文件「打开」按钮 → 文件选择框
}

// entryWidth 统一输入框固定宽度（所有文本输入行一致，不随窗口拉伸）。
const entryWidth = 300

// fixedWidthLayout 将子对象约束为固定宽度、占满可用高度。
// 必须用 Container+Layout 而非包装器 struct：包装器会经方法提升意外满足
// fyne.Widget 接口，被 Fyne 渲染树误当独立 widget 导致内容不渲染。
type fixedWidthLayout struct{ width float32 }

func (l fixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(fyne.NewSize(l.width, size.Height))
	}
}

func (l fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	h := float32(0)
	for _, o := range objects {
		if m := o.MinSize().Height; m > h {
			h = m
		}
	}
	return fyne.NewSize(l.width, h)
}

func newFixedWidth(obj fyne.CanvasObject, width float32) fyne.CanvasObject {
	return container.New(fixedWidthLayout{width: width}, obj)
}

// NewNinferForm 构建 9 组表单（垂直一排、带分组标题、整体滚动）并返回。
// 行布局统一：标签列宽取全部标签的最大文本宽度，输入框固定 entryWidth，
// 保证 34 个输入控件跨组左对齐、长度一致。
func NewNinferForm(window fyne.Window) *NinferForm {
	f := &NinferForm{
		window:  window,
		entries: make(map[string]*widget.Entry, len(fieldDefs)),
		checks:  make(map[string]*widget.Check),
		selects: make(map[string]*widget.Select),
	}

	controls := make(map[string]fyne.CanvasObject, len(fieldDefs))
	for _, d := range fieldDefs {
		switch d.kind {
		case kindBool:
			c := widget.NewCheck("", nil)
			c.OnChanged = func(bool) { f.notifyChange() }
			f.checks[d.name] = c
			controls[d.name] = c
		case kindSelect:
			s := widget.NewSelect(d.options, func(string) { f.notifyChange() })
			f.selects[d.name] = s
			controls[d.name] = s
		default:
			e := widget.NewEntry()
			e.OnChanged = func(string) { f.notifyChange() }
			f.entries[d.name] = e
			controls[d.name] = e
		}
	}

	// 统一标签列宽：取全部行标签（含「推理引擎」）的最大文本宽度。
	labelWidth := widget.NewLabel("推理引擎").MinSize().Width
	for _, d := range fieldDefs {
		if w := widget.NewLabel(d.label).MinSize().Width; w > labelWidth {
			labelWidth = w
		}
	}

	// row 构造统一行：固定宽标签 + 控件（文本输入固定宽，其余按自身宽度）+ 弹性空白。
	row := func(label string, w fyne.CanvasObject, fixed bool) fyne.CanvasObject {
		if fixed {
			w = newFixedWidth(w, entryWidth)
		}
		return container.NewHBox(newFixedWidth(widget.NewLabel(label), labelWidth), w, layout.NewSpacer())
	}

	// 9 组散开垂直排列：每组一个加粗标题 + 统一行；「服务」组后插入推理引擎行，
	// 「模型」组的 ModelPath 行带「打开」按钮。
	var sections []fyne.CanvasObject
	seen := make(map[string]bool, 8)
	for _, d := range fieldDefs {
		if !seen[d.group] {
			seen[d.group] = true
			sections = append(sections,
				widget.NewLabelWithStyle(d.group, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
			if d.group == "服务" {
				sections = append(sections, f.buildWorkerPathRow(labelWidth))
			}
		}
		if d.name == "ModelPath" {
			sections = append(sections, f.buildModelPathRow(labelWidth))
			continue
		}
		_, isCheck := f.checks[d.name]
		sections = append(sections, row(d.label, controls[d.name], !isCheck))
	}
	f.content = container.NewScroll(container.NewVBox(sections...))

	f.errLabel = widget.NewLabel("✓ 配置有效")
	f.errLabel.Wrapping = fyne.TextWrapWord
	// 构造期用默认配置初始化控件（Select 需要初始选中值，否则收集到空串）；
	// 此时 onChange 尚未接线，notifyChange 为 no-op。
	f.SetFieldsFrom(config.DefaultConfig())
	return f
}

// Content 返回表单滚动容器（供窗口布局使用）。
func (f *NinferForm) Content() fyne.CanvasObject {
	return f.content
}

// buildWorkerPathRow 构建「推理引擎」行：标签 + 固定宽路径输入框 + 打开按钮。
// 输入框回车提交路径；打开按钮弹文件选择框（回调由 App 注入）。
func (f *NinferForm) buildWorkerPathRow(labelWidth float32) fyne.CanvasObject {
	f.workerPathEntry = widget.NewEntry()
	f.workerPathEntry.SetPlaceHolder("NInfer 推理引擎路径（输入后回车，或点「打开」选择）")
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
		newFixedWidth(widget.NewLabel("推理引擎"), labelWidth),
		newFixedWidth(f.workerPathEntry, entryWidth),
		openBtn,
		layout.NewSpacer(),
	)
}

// buildModelPathRow 构建「模型文件」行：标签 + 固定宽路径输入框 + 打开按钮。
// 打开按钮弹文件选择框（回调由 App 注入），选择结果回填输入框（随配置保存）。
func (f *NinferForm) buildModelPathRow(labelWidth float32) fyne.CanvasObject {
	openBtn := widget.NewButton("打开", func() {
		if f.onOpenModel != nil {
			f.onOpenModel()
		}
	})
	return container.NewHBox(
		newFixedWidth(widget.NewLabel("模型文件（位置参数 model.ninfer）"), labelWidth),
		newFixedWidth(f.entries["ModelPath"], entryWidth),
		openBtn,
		layout.NewSpacer(),
	)
}

// SetWorkerPath 回填推理引擎路径输入框（主线程；SetText 不触发 OnSubmitted）。
func (f *NinferForm) SetWorkerPath(p string) {
	if f.workerPathEntry != nil {
		f.workerPathEntry.SetText(p)
	}
}

// SetModelPath 回填模型路径输入框（主线程；文件选择框回调使用）。
func (f *NinferForm) SetModelPath(p string) {
	if e, ok := f.entries["ModelPath"]; ok {
		e.SetText(p)
	}
}

// notifyChange 把字段变化转发给 App（主线程内立即返回）。
func (f *NinferForm) notifyChange() {
	if f.onChange != nil {
		f.onChange()
	}
}

// CountFields 返回字段总数（应为 34）。
func (f *NinferForm) CountFields() int { return len(fieldDefs) }

// controlCount 返回实际控件数（Entry + Check + Select，应为 34）。
func (f *NinferForm) controlCount() int {
	return len(f.entries) + len(f.checks) + len(f.selects)
}

// GroupNames 按字段出现顺序返回 8 个分组名。
func (f *NinferForm) GroupNames() []string {
	var out []string
	seen := map[string]bool{}
	for _, d := range fieldDefs {
		if !seen[d.group] {
			seen[d.group] = true
			out = append(out, d.group)
		}
	}
	return out
}

// SetFieldsFrom 用配置填充全部控件（主线程）。
func (f *NinferForm) SetFieldsFrom(cfg config.Config) {
	f.entries["ListenHost"].SetText(cfg.ListenHost)
	f.entries["Port"].SetText(strconv.Itoa(cfg.Port))
	f.entries["Concurrency"].SetText(strconv.Itoa(cfg.Concurrency))
	f.entries["ModelPath"].SetText(cfg.ModelPath)
	f.entries["Gpu"].SetText(cfg.Gpu)
	f.entries["StageLayers"].SetText(cfg.StageLayers)
	f.entries["Context"].SetText(strconv.Itoa(cfg.Context))
	f.entries["Budget"].SetText(strconv.Itoa(cfg.Budget))
	f.entries["Reserve"].SetText(strconv.Itoa(cfg.Reserve))
	f.entries["HostMiB"].SetText(strconv.Itoa(cfg.HostMiB))
	f.entries["MaxTokens"].SetText(strconv.Itoa(cfg.MaxTokens))
	f.entries["Prefill"].SetText(strconv.Itoa(cfg.Prefill))
	f.selects["KvType"].SetSelected(cfg.KvType)
	f.entries["SinkTokens"].SetText(strconv.Itoa(cfg.SinkTokens))
	f.entries["RecentTokens"].SetText(strconv.Itoa(cfg.RecentTokens))
	f.entries["RetainedSessions"].SetText(strconv.Itoa(cfg.RetainedSessions))
	f.checks["FastPrefill"].SetChecked(cfg.FastPrefill)
	f.entries["MtpDrafts"].SetText(strconv.Itoa(cfg.MtpDrafts))
	f.entries["NgramDrafts"].SetText(strconv.Itoa(cfg.NgramDrafts))
	f.entries["NgramMinMatch"].SetText(strconv.Itoa(cfg.NgramMinMatch))
	f.checks["AdaptiveMtp"].SetChecked(cfg.AdaptiveMtp)
	f.checks["Vision"].SetChecked(cfg.Vision)
	f.selects["VisionResidency"].SetSelected(cfg.VisionResidency)
	f.entries["VisionTokens"].SetText(strconv.Itoa(cfg.VisionTokens))
	f.entries["ThinkingBudget"].SetText(strconv.Itoa(cfg.ThinkingBudget))
	f.selects["DefaultReasoningEffort"].SetSelected(cfg.DefaultReasoningEffort)
	f.selects["DeviceProfile"].SetSelected(cfg.DeviceProfile)
	f.entries["DiskPath"].SetText(cfg.DiskPath)
	f.entries["DiskMiB"].SetText(strconv.Itoa(cfg.DiskMiB))
	f.checks["NoLrfu"].SetChecked(cfg.NoLrfu)
	f.entries["LrfuLambda"].SetText(strconv.FormatFloat(cfg.LrfuLambda, 'f', -1, 64))
	f.entries["LrfuTopFraction"].SetText(strconv.FormatFloat(cfg.LrfuTopFraction, 'f', -1, 64))
	f.selects["ScoreAggregation"].SetSelected(cfg.ScoreAggregation)
}

// CollectConfig 从控件收集当前配置（主线程轻量读）。
// 第二个返回值是中文解析错误列表（空串按默认值处理）。
func (f *NinferForm) CollectConfig() (config.Config, []string) {
	var cfg config.Config
	var errs []string

	cfg.ListenHost = f.entries["ListenHost"].Text
	cfg.Port, errs = collectInt(f.entries["Port"], "监听端口", cfg.Port, errs)
	cfg.Concurrency, errs = collectInt(f.entries["Concurrency"], "并发解码请求数", cfg.Concurrency, errs)
	cfg.ModelPath = f.entries["ModelPath"].Text
	cfg.Gpu = f.entries["Gpu"].Text
	cfg.StageLayers = f.entries["StageLayers"].Text
	cfg.Context, errs = collectInt(f.entries["Context"], "上下文上限", cfg.Context, errs)
	cfg.Budget, errs = collectInt(f.entries["Budget"], "历史预算", cfg.Budget, errs)
	cfg.Reserve, errs = collectInt(f.entries["Reserve"], "生成保留", cfg.Reserve, errs)
	cfg.HostMiB, errs = collectInt(f.entries["HostMiB"], "Host KV 预算", cfg.HostMiB, errs)
	cfg.MaxTokens, errs = collectInt(f.entries["MaxTokens"], "默认输出上限", cfg.MaxTokens, errs)
	cfg.Prefill, errs = collectInt(f.entries["Prefill"], "prefill 块", cfg.Prefill, errs)
	cfg.KvType = f.selects["KvType"].Selected
	cfg.SinkTokens, errs = collectInt(f.entries["SinkTokens"], "常驻前缀", cfg.SinkTokens, errs)
	cfg.RecentTokens, errs = collectInt(f.entries["RecentTokens"], "常驻后缀", cfg.RecentTokens, errs)
	cfg.RetainedSessions, errs = collectInt(f.entries["RetainedSessions"], "保留历史数", cfg.RetainedSessions, errs)
	cfg.FastPrefill = f.checks["FastPrefill"].Checked
	cfg.MtpDrafts, errs = collectInt(f.entries["MtpDrafts"], "MTP 草稿数", cfg.MtpDrafts, errs)
	cfg.NgramDrafts, errs = collectInt(f.entries["NgramDrafts"], "ngram 草稿数", cfg.NgramDrafts, errs)
	cfg.NgramMinMatch, errs = collectInt(f.entries["NgramMinMatch"], "ngram 最短匹配", cfg.NgramMinMatch, errs)
	cfg.AdaptiveMtp = f.checks["AdaptiveMtp"].Checked
	cfg.Vision = f.checks["Vision"].Checked
	cfg.VisionResidency = f.selects["VisionResidency"].Selected
	cfg.VisionTokens, errs = collectInt(f.entries["VisionTokens"], "视觉合并上限", cfg.VisionTokens, errs)
	cfg.ThinkingBudget, errs = collectInt(f.entries["ThinkingBudget"], "思考预算", cfg.ThinkingBudget, errs)
	cfg.DefaultReasoningEffort = f.selects["DefaultReasoningEffort"].Selected
	cfg.DeviceProfile = f.selects["DeviceProfile"].Selected
	cfg.DiskPath = f.entries["DiskPath"].Text
	cfg.DiskMiB, errs = collectInt(f.entries["DiskMiB"], "SSD 配额", cfg.DiskMiB, errs)
	cfg.NoLrfu = f.checks["NoLrfu"].Checked
	cfg.LrfuLambda, errs = collectFloat(f.entries["LrfuLambda"], "历史衰减", cfg.LrfuLambda, errs)
	cfg.LrfuTopFraction, errs = collectFloat(f.entries["LrfuTopFraction"], "命中比例", cfg.LrfuTopFraction, errs)
	cfg.ScoreAggregation = f.selects["ScoreAggregation"].Selected
	return cfg, errs
}

func collectInt(e *widget.Entry, label string, def int, errs []string) (int, []string) {
	s := strings.TrimSpace(e.Text)
	if s == "" {
		return def, errs
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def, append(errs, fmt.Sprintf("%s 必须是整数（当前 %q）", label, s))
	}
	return v, errs
}

func collectFloat(e *widget.Entry, label string, def float64, errs []string) (float64, []string) {
	s := strings.TrimSpace(e.Text)
	if s == "" {
		return def, errs
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def, append(errs, fmt.Sprintf("%s 必须是数字（当前 %q）", label, s))
	}
	return v, errs
}

// SetValidationErrors 显示校验结果（中文，每行一条；主线程）。
// config.Validate 直接返回中文文案，此处只做展示。
func (f *NinferForm) SetValidationErrors(errs []string) {
	if len(errs) == 0 {
		f.errLabel.SetText("✓ 配置有效")
		return
	}
	var b strings.Builder
	for i, e := range errs {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "✗ %s", e)
	}
	f.errLabel.SetText(b.String())
}
