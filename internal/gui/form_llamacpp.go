// form_llamacpp.go llama.cpp 引擎的表单（46 字段 / 9 组）。
//
// 与 ninfer 表单（form_ninfer.go）完全分离：字段集、默认值、argv flag
// 都不同（对应 Qwen3.8-27B-GSQ_S-MTP.cmd 启动脚本）。复用同包的
// newFixedWidth/entryWidth/collectInt/collectFloat 布局与解析工具。
package gui

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"llmlauncher/internal/config"
)

// llamaFieldDef 是 llama 表单对单个字段的唯一声明。
type llamaFieldDef struct {
	group   string
	name    string
	label   string
	kind    fieldKind
	options []string // kindSelect 的选项
}

// llamaFieldDefs：46 字段，9 组，顺序与 config.LlamaConfig 字段顺序一致。
// 字段与默认值对齐 Qwen3.8-27B-GSQ_S-MTP.cmd 启动脚本（真实 llama-server 参数）；
// 标签末尾括号内为对应命令行参数。
var llamaFieldDefs = []llamaFieldDef{
	// 服务（4）
	{"服务", "Host", "监听地址（--host）", kindString, nil},
	{"服务", "Port", "监听端口（--port）", kindInt, nil},
	{"服务", "Threads", "线程数（-t）", kindInt, nil},
	{"服务", "Parallel", "并行请求数（--parallel）", kindInt, nil},
	// 模型（7）
	{"模型", "ModelPath", "模型文件（GGUF 路径，-m）", kindString, nil},
	{"模型", "MmprojPath", "视觉模型（--mmproj，空 = 不传）", kindString, nil},
	{"模型", "MmprojOffload", "视觉模型卸载（--mmproj-offload）", kindBool, nil},
	{"模型", "ImageMinTokens", "图像最少 token（--image-min-tokens）", kindInt, nil},
	{"模型", "ImageMaxTokens", "图像最多 token（--image-max-tokens）", kindInt, nil},
	{"模型", "Alias", "模型别名（--alias）", kindString, nil},
	{"模型", "ChatTemplateFile", "聊天模板文件（--chat-template-file，空 = 不传）", kindString, nil},
	// 上下文与批处理（5）
	{"上下文与批处理", "CtxSize", "上下文长度（-c）", kindInt, nil},
	{"上下文与批处理", "NPredict", "最大回复长度（-n）", kindInt, nil},
	{"上下文与批处理", "BatchSize", "批大小（-b）", kindInt, nil},
	{"上下文与批处理", "UbatchSize", "微批大小（-ub）", kindInt, nil},
	{"上下文与批处理", "Keep", "保留 token 数（--keep，-1 = 全保留）", kindInt, nil},
	// GPU 与加载（4）
	{"GPU 与加载", "GpuLayers", "GPU 层数（--gpu-layers）", kindInt, nil},
	{"GPU 与加载", "GpuLayersDraft", "草稿模型 GPU 层数（--gpu-layers-draft）", kindString, nil},
	{"GPU 与加载", "LoadMode", "加载模式（--load-mode）", kindString, nil},
	{"GPU 与加载", "NoWarmup", "跳过预热（--no-warmup）", kindBool, nil},
	// KV 缓存（7）
	{"KV 缓存", "CacheTypeV", "V 缓存量化（--cache-type-v）", kindSelect,
		[]string{"f16", "bf16", "q8_0", "q4_0", "q4_1", "iq4_nl"}},
	{"KV 缓存", "CacheTypeK", "K 缓存量化（--cache-type-k）", kindSelect,
		[]string{"f16", "bf16", "q8_0", "q4_0", "q4_1", "iq4_nl"}},
	{"KV 缓存", "CacheTypeVDraft", "草稿 V 缓存量化（--cache-type-v-draft）", kindSelect,
		[]string{"f16", "bf16", "q8_0", "q4_0", "q4_1", "iq4_nl"}},
	{"KV 缓存", "CacheTypeKDraft", "草稿 K 缓存量化（--cache-type-k-draft）", kindSelect,
		[]string{"f16", "bf16", "q8_0", "q4_0", "q4_1", "iq4_nl"}},
	{"KV 缓存", "CacheRam", "缓存内存 MiB（--cache-ram）", kindInt, nil},
	{"KV 缓存", "NoCacheIdleSlots", "禁用空闲槽（--no-cache-idle-slots）", kindBool, nil},
	{"KV 缓存", "NoHost", "禁用 host 缓存（--no-host）", kindBool, nil},
	// 推测解码（4）
	{"推测解码", "SpecType", "推测类型（--spec-type）", kindString, nil},
	{"推测解码", "SpecDraftNMin", "草稿最小长度（--spec-draft-n-min）", kindInt, nil},
	{"推测解码", "SpecDraftNMax", "草稿最大长度（--spec-draft-n-max）", kindInt, nil},
	{"推测解码", "KvUnified", "统一 KV（--kv-unified）", kindBool, nil},
	// 注意力与检查点（3）
	{"注意力与检查点", "FlashAttn", "FlashAttention（-fa）", kindSelect,
		[]string{"on", "off"}},
	{"注意力与检查点", "CtxCheckpoints", "上下文检查点数（--ctx-checkpoints）", kindInt, nil},
	{"注意力与检查点", "Poll", "轮询间隔（--poll）", kindInt, nil},
	// 推理行为（6）
	{"推理行为", "Jinja", "启用 Jinja 模板（--jinja）", kindBool, nil},
	{"推理行为", "Reasoning", "思考模式（--reasoning）", kindSelect,
		[]string{"on", "off"}},
	{"推理行为", "ReasoningPreserve", "保留思考内容（--reasoning-preserve）", kindBool, nil},
	{"推理行为", "ReasoningFormat", "思考格式（--reasoning-format）", kindSelect,
		[]string{"none", "deepseek"}},
	{"推理行为", "Fit", "上下文适配（-fit）", kindSelect,
		[]string{"on", "off"}},
	{"推理行为", "ChatTemplateKwargs", "模板参数（--chat-template-kwargs，JSON，空 = 不传）", kindString, nil},
	// 采样（6）
	{"采样", "Temp", "温度（--temp）", kindFloat, nil},
	{"采样", "TopP", "top-p（(0,1]）", kindFloat, nil},
	{"采样", "TopK", "top-k", kindInt, nil},
	{"采样", "MinP", "min-p（[0,1]）", kindFloat, nil},
	{"采样", "PresencePenalty", "存在惩罚（--presence-penalty）", kindFloat, nil},
	{"采样", "RepeatPenalty", "重复惩罚（--repeat-penalty）", kindFloat, nil},
}

// LlamaForm 持有 llama.cpp 表单全部控件（46 个）与校验状态行。
// 所有方法必须运行在主线程（Fyne 约束）。
type LlamaForm struct {
	window   fyne.Window
	content  fyne.CanvasObject
	entries  map[string]*widget.Entry
	checks   map[string]*widget.Check
	selects  map[string]*widget.Select
	errLabel *widget.Label

	workerPathEntry *widget.Entry // llama worker 可执行文件路径输入框

	onChange        func()            // App 注入：字段变化时触发后台校验
	onOpenWorker    func()            // App 注入：打开按钮 → 文件选择窗口
	onWorkerPathSet func(path string) // App 注入：输入框回车提交路径
	onOpenModel     func()            // App 注入：模型文件「打开」按钮 → 文件选择窗口
	onOpenMmproj    func()            // App 注入：视觉模型「打开」按钮 → 文件选择窗口
	onOpenChatTpl   func()            // App 注入：聊天模板文件「打开」按钮 → 文件选择窗口
}

// NewLlamaForm 构建 9 组表单（垂直一排、带分组标题、整体滚动）并返回。
// 行布局与 ninfer 表单一致：统一标签列宽 + 固定输入框宽度。
func NewLlamaForm(window fyne.Window) *LlamaForm {
	f := &LlamaForm{
		window:  window,
		entries: make(map[string]*widget.Entry, len(llamaFieldDefs)),
		checks:  make(map[string]*widget.Check),
		selects: make(map[string]*widget.Select),
	}

	controls := make(map[string]fyne.CanvasObject, len(llamaFieldDefs))
	for _, d := range llamaFieldDefs {
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
	seen := make(map[string]bool, 9)
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
		_, isSelect := f.selects[d.name]
		if d.name == "ModelPath" {
			sections = append(sections, f.buildModelPathRow(labelWidth))
			continue
		}
		if d.name == "MmprojPath" {
			sections = append(sections, f.buildMmprojPathRow(labelWidth))
			continue
		}
		if d.name == "ChatTemplateFile" {
			sections = append(sections, f.buildChatTemplatePathRow(labelWidth))
			continue
		}
		sections = append(sections, row(d.label, controls[d.name], !isCheck && !isSelect))
	}
	f.content = container.NewScroll(container.NewVBox(sections...))

	f.errLabel = widget.NewLabel("✓ 配置有效")
	f.errLabel.Wrapping = fyne.TextWrapWord
	// 构造期用默认配置初始化控件（Select 需要初始选中值，否则收集到空串）；
	// 此时 onChange 尚未接线，notifyChange 为 no-op。
	f.SetFieldsFrom(config.DefaultLlamaConfig())
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

// buildModelPathRow 构建「模型文件」行：标签 + 固定宽路径输入框 + 打开按钮。
// 打开按钮弹文件选择窗口（回调由 App 注入），选择结果回填输入框（随配置保存）。
func (f *LlamaForm) buildModelPathRow(labelWidth float32) fyne.CanvasObject {
	openBtn := widget.NewButton("打开", func() {
		if f.onOpenModel != nil {
			f.onOpenModel()
		}
	})
	return container.NewHBox(
		newFixedWidth(widget.NewLabel("模型文件（GGUF 路径，-m）"), labelWidth),
		newFixedWidth(f.entries["ModelPath"], entryWidth),
		openBtn,
		layout.NewSpacer(),
	)
}

// buildMmprojPathRow 构建「视觉模型」行：标签 + 固定宽路径输入框 + 打开按钮。
func (f *LlamaForm) buildMmprojPathRow(labelWidth float32) fyne.CanvasObject {
	openBtn := widget.NewButton("打开", func() {
		if f.onOpenMmproj != nil {
			f.onOpenMmproj()
		}
	})
	return container.NewHBox(
		newFixedWidth(widget.NewLabel("视觉模型（--mmproj，空 = 不传）"), labelWidth),
		newFixedWidth(f.entries["MmprojPath"], entryWidth),
		openBtn,
		layout.NewSpacer(),
	)
}

// SetModelPath 回填模型路径输入框（主线程；文件选择窗口回调使用）。
func (f *LlamaForm) SetModelPath(p string) {
	if e, ok := f.entries["ModelPath"]; ok {
		e.SetText(p)
	}
}

// SetMmprojPath 回填视觉模型路径输入框（主线程；文件选择窗口回调使用）。
func (f *LlamaForm) SetMmprojPath(p string) {
	if e, ok := f.entries["MmprojPath"]; ok {
		e.SetText(p)
	}
}

// buildChatTemplatePathRow 构建「聊天模板文件」行：标签 + 固定宽路径输入框 + 打开按钮。
func (f *LlamaForm) buildChatTemplatePathRow(labelWidth float32) fyne.CanvasObject {
	openBtn := widget.NewButton("打开", func() {
		if f.onOpenChatTpl != nil {
			f.onOpenChatTpl()
		}
	})
	return container.NewHBox(
		newFixedWidth(widget.NewLabel("聊天模板文件（--chat-template-file，空 = 不传）"), labelWidth),
		newFixedWidth(f.entries["ChatTemplateFile"], entryWidth),
		openBtn,
		layout.NewSpacer(),
	)
}

// SetChatTemplatePath 回填聊天模板文件路径输入框（主线程；文件选择窗口回调使用）。
func (f *LlamaForm) SetChatTemplatePath(p string) {
	if e, ok := f.entries["ChatTemplateFile"]; ok {
		e.SetText(p)
	}
}

// notifyChange 把字段变化转发给 App（主线程内立即返回）。
func (f *LlamaForm) notifyChange() {
	if f.onChange != nil {
		f.onChange()
	}
}

// CountFields 返回字段总数（应为 46）。
func (f *LlamaForm) CountFields() int { return len(llamaFieldDefs) }

// GroupNames 返回分组名列表（去重、按出现顺序）。
func (f *LlamaForm) GroupNames() []string {
	var names []string
	seen := make(map[string]bool, 9)
	for _, d := range llamaFieldDefs {
		if !seen[d.group] {
			seen[d.group] = true
			names = append(names, d.group)
		}
	}
	return names
}

// SetFieldsFrom 用配置填充全部控件（主线程）。
func (f *LlamaForm) SetFieldsFrom(cfg config.LlamaConfig) {
	f.entries["Host"].SetText(cfg.Host)
	f.entries["Port"].SetText(strconv.Itoa(cfg.Port))
	f.entries["Threads"].SetText(strconv.Itoa(cfg.Threads))
	f.entries["Parallel"].SetText(strconv.Itoa(cfg.Parallel))

	f.entries["ModelPath"].SetText(cfg.ModelPath)
	f.entries["MmprojPath"].SetText(cfg.MmprojPath)
	f.checks["MmprojOffload"].SetChecked(cfg.MmprojOffload)
	f.entries["ImageMinTokens"].SetText(strconv.Itoa(cfg.ImageMinTokens))
	f.entries["ImageMaxTokens"].SetText(strconv.Itoa(cfg.ImageMaxTokens))
	f.entries["Alias"].SetText(cfg.Alias)
	f.entries["ChatTemplateFile"].SetText(cfg.ChatTemplateFile)

	f.entries["CtxSize"].SetText(strconv.Itoa(cfg.CtxSize))
	f.entries["NPredict"].SetText(strconv.Itoa(cfg.NPredict))
	f.entries["BatchSize"].SetText(strconv.Itoa(cfg.BatchSize))
	f.entries["UbatchSize"].SetText(strconv.Itoa(cfg.UbatchSize))
	f.entries["Keep"].SetText(strconv.Itoa(cfg.Keep))

	f.entries["GpuLayers"].SetText(strconv.Itoa(cfg.GpuLayers))
	f.entries["GpuLayersDraft"].SetText(cfg.GpuLayersDraft)
	f.entries["LoadMode"].SetText(cfg.LoadMode)
	f.checks["NoWarmup"].SetChecked(cfg.NoWarmup)

	f.selects["CacheTypeV"].SetSelected(cfg.CacheTypeV)
	f.selects["CacheTypeK"].SetSelected(cfg.CacheTypeK)
	f.selects["CacheTypeVDraft"].SetSelected(cfg.CacheTypeVDraft)
	f.selects["CacheTypeKDraft"].SetSelected(cfg.CacheTypeKDraft)
	f.entries["CacheRam"].SetText(strconv.Itoa(cfg.CacheRam))
	f.checks["NoCacheIdleSlots"].SetChecked(cfg.NoCacheIdleSlots)
	f.checks["NoHost"].SetChecked(cfg.NoHost)

	f.entries["SpecType"].SetText(cfg.SpecType)
	f.entries["SpecDraftNMin"].SetText(strconv.Itoa(cfg.SpecDraftNMin))
	f.entries["SpecDraftNMax"].SetText(strconv.Itoa(cfg.SpecDraftNMax))
	f.checks["KvUnified"].SetChecked(cfg.KvUnified)

	f.selects["FlashAttn"].SetSelected(cfg.FlashAttn)
	f.entries["CtxCheckpoints"].SetText(strconv.Itoa(cfg.CtxCheckpoints))
	f.entries["Poll"].SetText(strconv.Itoa(cfg.Poll))

	f.checks["Jinja"].SetChecked(cfg.Jinja)
	f.selects["Reasoning"].SetSelected(cfg.Reasoning)
	f.checks["ReasoningPreserve"].SetChecked(cfg.ReasoningPreserve)
	f.selects["ReasoningFormat"].SetSelected(cfg.ReasoningFormat)
	f.selects["Fit"].SetSelected(cfg.Fit)
	f.entries["ChatTemplateKwargs"].SetText(cfg.ChatTemplateKwargs)

	f.entries["Temp"].SetText(strconv.FormatFloat(cfg.Temp, 'f', -1, 64))
	f.entries["TopP"].SetText(strconv.FormatFloat(cfg.TopP, 'f', -1, 64))
	f.entries["TopK"].SetText(strconv.Itoa(cfg.TopK))
	f.entries["MinP"].SetText(strconv.FormatFloat(cfg.MinP, 'f', -1, 64))
	f.entries["PresencePenalty"].SetText(strconv.FormatFloat(cfg.PresencePenalty, 'f', -1, 64))
	f.entries["RepeatPenalty"].SetText(strconv.FormatFloat(cfg.RepeatPenalty, 'f', -1, 64))
}

// CollectConfig 从控件收集当前配置（主线程轻量读）。
// 第二个返回值是中文解析错误列表（空串按默认值处理）。
func (f *LlamaForm) CollectConfig() (config.LlamaConfig, []string) {
	var cfg config.LlamaConfig
	var errs []string

	cfg.Host = f.entries["Host"].Text
	cfg.Port, errs = collectInt(f.entries["Port"], "监听端口", cfg.Port, errs)
	cfg.Threads, errs = collectInt(f.entries["Threads"], "线程数", cfg.Threads, errs)
	cfg.Parallel, errs = collectInt(f.entries["Parallel"], "并行请求数", cfg.Parallel, errs)

	cfg.ModelPath = f.entries["ModelPath"].Text
	cfg.MmprojPath = f.entries["MmprojPath"].Text
	cfg.MmprojOffload = f.checks["MmprojOffload"].Checked
	cfg.ImageMinTokens, errs = collectInt(f.entries["ImageMinTokens"], "图像最少 token", cfg.ImageMinTokens, errs)
	cfg.ImageMaxTokens, errs = collectInt(f.entries["ImageMaxTokens"], "图像最多 token", cfg.ImageMaxTokens, errs)
	cfg.Alias = f.entries["Alias"].Text
	cfg.ChatTemplateFile = f.entries["ChatTemplateFile"].Text

	cfg.CtxSize, errs = collectInt(f.entries["CtxSize"], "上下文长度", cfg.CtxSize, errs)
	cfg.NPredict, errs = collectInt(f.entries["NPredict"], "最大回复长度", cfg.NPredict, errs)
	cfg.BatchSize, errs = collectInt(f.entries["BatchSize"], "批大小", cfg.BatchSize, errs)
	cfg.UbatchSize, errs = collectInt(f.entries["UbatchSize"], "微批大小", cfg.UbatchSize, errs)
	cfg.Keep, errs = collectInt(f.entries["Keep"], "保留 token 数", cfg.Keep, errs)

	cfg.GpuLayers, errs = collectInt(f.entries["GpuLayers"], "GPU 层数", cfg.GpuLayers, errs)
	cfg.GpuLayersDraft = f.entries["GpuLayersDraft"].Text
	cfg.LoadMode = f.entries["LoadMode"].Text
	cfg.NoWarmup = f.checks["NoWarmup"].Checked

	cfg.CacheTypeV = f.selects["CacheTypeV"].Selected
	cfg.CacheTypeK = f.selects["CacheTypeK"].Selected
	cfg.CacheTypeVDraft = f.selects["CacheTypeVDraft"].Selected
	cfg.CacheTypeKDraft = f.selects["CacheTypeKDraft"].Selected
	cfg.CacheRam, errs = collectInt(f.entries["CacheRam"], "缓存内存", cfg.CacheRam, errs)
	cfg.NoCacheIdleSlots = f.checks["NoCacheIdleSlots"].Checked
	cfg.NoHost = f.checks["NoHost"].Checked

	cfg.SpecType = f.entries["SpecType"].Text
	cfg.SpecDraftNMin, errs = collectInt(f.entries["SpecDraftNMin"], "草稿最小长度", cfg.SpecDraftNMin, errs)
	cfg.SpecDraftNMax, errs = collectInt(f.entries["SpecDraftNMax"], "草稿最大长度", cfg.SpecDraftNMax, errs)
	cfg.KvUnified = f.checks["KvUnified"].Checked

	cfg.FlashAttn = f.selects["FlashAttn"].Selected
	cfg.CtxCheckpoints, errs = collectInt(f.entries["CtxCheckpoints"], "上下文检查点数", cfg.CtxCheckpoints, errs)
	cfg.Poll, errs = collectInt(f.entries["Poll"], "轮询间隔", cfg.Poll, errs)

	cfg.Jinja = f.checks["Jinja"].Checked
	cfg.Reasoning = f.selects["Reasoning"].Selected
	cfg.ReasoningPreserve = f.checks["ReasoningPreserve"].Checked
	cfg.ReasoningFormat = f.selects["ReasoningFormat"].Selected
	cfg.Fit = f.selects["Fit"].Selected
	cfg.ChatTemplateKwargs = f.entries["ChatTemplateKwargs"].Text

	cfg.Temp, errs = collectFloat(f.entries["Temp"], "温度", cfg.Temp, errs)
	cfg.TopP, errs = collectFloat(f.entries["TopP"], "top-p", cfg.TopP, errs)
	cfg.TopK, errs = collectInt(f.entries["TopK"], "top-k", cfg.TopK, errs)
	cfg.MinP, errs = collectFloat(f.entries["MinP"], "min-p", cfg.MinP, errs)
	cfg.PresencePenalty, errs = collectFloat(f.entries["PresencePenalty"], "存在惩罚", cfg.PresencePenalty, errs)
	cfg.RepeatPenalty, errs = collectFloat(f.entries["RepeatPenalty"], "重复惩罚", cfg.RepeatPenalty, errs)
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
