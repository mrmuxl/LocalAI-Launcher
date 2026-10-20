// form.go 构建 34 字段 / 8 分组的启动器表单，并承载校验、保存、加载交互。
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
)

// fieldDef 是表单对单个字段的唯一声明。
type fieldDef struct {
	group string
	name  string
	label string
	kind  fieldKind
}

// fieldDefs：34 字段，8 组，顺序与 config.Config 字段顺序一致。
// 标签末尾括号内为对应命令行参数（来源：docs/命令行参数参考.txt 与 argv 构造契约）。
var fieldDefs = []fieldDef{
	// 服务（4）
	{"服务", "Host", "监听地址（--bind）", kindString},
	{"服务", "Port", "监听端口（--port）", kindInt},
	{"服务", "LogLevel", "日志级别（--log-level，debug/info/warn/error）", kindString},
	{"服务", "Workers", "并发 worker 数（--workers）", kindInt},
	// 模型（5）
	{"模型", "ModelName", "模型标识（--model）", kindString},
	{"模型", "ModelPath", "模型文件/目录（位置参数 model.ninfer，空 = 自动）", kindString},
	{"模型", "CtxSize", "上下文长度（--ctx）", kindInt},
	{"模型", "Threads", "推理线程（--threads，0 = 自动）", kindInt},
	{"模型", "Quant", "量化格式（--quant）", kindString},
	// 采样（5）
	{"采样", "Temperature", "采样温度（--temp，0..2）", kindFloat},
	{"采样", "TopP", "top-p（--top-p，0..1）", kindFloat},
	{"采样", "TopK", "top-k（--top-k，0..20）", kindInt},
	{"采样", "RepeatPenalty", "重复惩罚（--repeat）", kindFloat},
	{"采样", "MaxTokens", "最大生成 token（--max-tokens）", kindInt},
	// 网络（4）
	{"网络", "APIKey", "鉴权密钥（--api-key，可选）", kindString},
	{"网络", "CORSOrigins", "允许来源（--cors，逗号分隔）", kindString},
	{"网络", "TimeoutSec", "请求超时秒数（--timeout）", kindInt},
	{"网络", "MaxConns", "最大并发连接（--max-conns）", kindInt},
	// 安全（3）
	{"安全", "TLS", "启用 TLS（--tls）", kindBool},
	{"安全", "CertPath", "证书路径（--cert）", kindString},
	{"安全", "KeyPath", "私钥路径（--key）", kindString},
	// 观测（4）
	{"观测", "MetricsEnabled", "启用 /metrics（--metrics）", kindBool},
	{"观测", "MetricsPath", "指标路径（--metrics-path）", kindString},
	{"观测", "LogFile", "日志文件（--log-file，空 = stdout）", kindString},
	{"观测", "AccessLog", "访问日志（暂未接入命令行）", kindBool},
	// 资源（5）
	{"资源", "MemLimitGB", "内存上限 GB（--mem-limit-gb，0 = 不限）", kindInt},
	{"资源", "SwapGB", "swap 上限 GB（--swap-gb）", kindInt},
	{"资源", "NumaNode", "NUMA 节点（--numa，-1 = 自动）", kindInt},
	{"资源", "Priority", "进程优先级（--priority，low/normal/high/realtime）", kindString},
	{"资源", "AffinityMask", "CPU 亲和掩码（--affinity，空 = 默认）", kindString},
	// 缓存（4）
	{"缓存", "CacheDir", "缓存目录（--cache-dir）", kindString},
	{"缓存", "CacheSizeMB", "缓存大小 MB（--cache-mb）", kindInt},
	{"缓存", "EvictPolicy", "淘汰策略（--evict，lru/fifo）", kindString},
	{"缓存", "Prefetch", "预取开关（--prefetch）", kindBool},
}

// validationZh 把 config.Validate 返回的「错误码:参数」串映射为中文文案。
// 键为错误码前缀；含 %s 的模板填入冒号后的参数值。
var validationZh = map[string]string{
	"V1:Port=":                             "监听端口 %s 超出范围 [1,65535]",
	"V1:Workers=":                          "并发 worker 数 %s 超出范围 [1,256]",
	"V1:CtxSize=":                          "上下文长度 %s 超出范围 [1,1048576]",
	"V2:MaxTokens=":                        "最大生成 token %s 必须 ≥ 1",
	"V2:TimeoutSec=":                       "请求超时 %s 必须 ≥ 1",
	"V2:MaxConns=":                         "最大并发连接 %s 超出范围 [1,1024]",
	"V3:Temperature=":                      "采样温度 %s 必须 ≥ 0",
	"V3:TopP=":                             "top-p %s 超出范围 [0,1]",
	"V3:TopK=":                             "top-k %s 必须 ≥ 0",
	"V3:RepeatPenalty=":                    "重复惩罚 %s 必须 > 0",
	"V4:TLS enabled but CertPath is empty": "已启用 TLS 但未填写证书路径",
	"V4:cert file not found: ":             "证书文件不存在：%s",
	"V4:TLS enabled but KeyPath is empty":  "已启用 TLS 但未填写私钥路径",
	"V4:key file not found: ":              "私钥文件不存在：%s",
	"V5:LogLevel=":                         "日志级别 %s 无效（应为 debug/info/warn/error 之一）",
	"V5:Priority=":                         "进程优先级 %s 无效（应为 low/normal/high/realtime 之一）",
	"V5:EvictPolicy=":                      "淘汰策略 %s 无效（应为 lru/fifo 之一）",
	"V6:MemLimitGB=":                       "内存上限 %s 必须 ≥ 0",
	"V6:SwapGB=":                           "swap 上限 %s 必须 ≥ 0",
	"V6:NumaNode=":                         "NUMA 节点 %s 必须为 -1 或 ≥ 0",
	"V6:CacheSizeMB=":                      "缓存大小 %s 必须 ≥ 1",
	"V7:path contains NUL byte":            "路径字段包含非法字符（NUL）",
}

// NinferForm 持有全部表单控件（34 个）与校验状态行。
// 所有方法必须运行在主线程（Fyne 约束）。
type NinferForm struct {
	window   fyne.Window
	content  fyne.CanvasObject // 8 组表单垂直排列的滚动容器
	entries  map[string]*widget.Entry
	checks   map[string]*widget.Check
	errLabel *widget.Label
	saveBtn  *widget.Button
	loadBtn  *widget.Button

	workerPathEntry *widget.Entry // 推理程序路径输入框（服务组下方）

	onChange        func()            // App 注入：字段变化时触发后台校验
	onOpenWorker    func()            // App 注入：打开按钮 → 文件选择框
	onWorkerPathSet func(path string) // App 注入：输入框回车提交路径
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

// NewNinferForm 构建 8 组表单（垂直一排、带分组标题、整体滚动）并返回。
// 行布局统一：标签列宽取全部标签的最大文本宽度，输入框固定 entryWidth，
// 保证 34 个输入控件跨组左对齐、长度一致。
func NewNinferForm(window fyne.Window) *NinferForm {
	f := &NinferForm{
		window:  window,
		entries: make(map[string]*widget.Entry, len(fieldDefs)),
		checks:  make(map[string]*widget.Check),
	}

	controls := make(map[string]fyne.CanvasObject, len(fieldDefs))
	for _, d := range fieldDefs {
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

	// 8 组散开垂直排列：每组一个加粗标题 + 统一行；「服务」组后插入推理程序行。
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
		_, isCheck := f.checks[d.name]
		sections = append(sections, row(d.label, controls[d.name], !isCheck))
	}
	f.content = container.NewScroll(container.NewVBox(sections...))

	f.errLabel = widget.NewLabel("✓ 配置有效")
	f.errLabel.Wrapping = fyne.TextWrapWord
	f.saveBtn = widget.NewButton("保存配置", nil)
	f.loadBtn = widget.NewButton("重新加载", nil)
	return f
}

// Content 返回表单滚动容器（供窗口布局使用）。
func (f *NinferForm) Content() fyne.CanvasObject {
	return f.content
}

// buildWorkerPathRow 构建「推理程序」行：标签 + 固定宽路径输入框 + 打开按钮。
// 输入框回车提交路径；打开按钮弹文件选择框（回调由 App 注入）。
func (f *NinferForm) buildWorkerPathRow(labelWidth float32) fyne.CanvasObject {
	f.workerPathEntry = widget.NewEntry()
	f.workerPathEntry.SetPlaceHolder("ninfer.exe 路径（输入后回车，或点「打开」选择）")
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
func (f *NinferForm) SetWorkerPath(p string) {
	if f.workerPathEntry != nil {
		f.workerPathEntry.SetText(p)
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
	f.entries["Host"].SetText(cfg.Host)
	f.entries["Port"].SetText(strconv.Itoa(cfg.Port))
	f.entries["LogLevel"].SetText(cfg.LogLevel)
	f.entries["Workers"].SetText(strconv.Itoa(cfg.Workers))
	f.entries["ModelName"].SetText(cfg.ModelName)
	f.entries["ModelPath"].SetText(cfg.ModelPath)
	f.entries["CtxSize"].SetText(strconv.Itoa(cfg.CtxSize))
	f.entries["Threads"].SetText(strconv.Itoa(cfg.Threads))
	f.entries["Quant"].SetText(cfg.Quant)
	f.entries["Temperature"].SetText(strconv.FormatFloat(cfg.Temperature, 'f', -1, 64))
	f.entries["TopP"].SetText(strconv.FormatFloat(cfg.TopP, 'f', -1, 64))
	f.entries["TopK"].SetText(strconv.Itoa(cfg.TopK))
	f.entries["RepeatPenalty"].SetText(strconv.FormatFloat(cfg.RepeatPenalty, 'f', -1, 64))
	f.entries["MaxTokens"].SetText(strconv.Itoa(cfg.MaxTokens))
	f.entries["APIKey"].SetText(cfg.APIKey)
	f.entries["CORSOrigins"].SetText(cfg.CORSOrigins)
	f.entries["TimeoutSec"].SetText(strconv.Itoa(cfg.TimeoutSec))
	f.entries["MaxConns"].SetText(strconv.Itoa(cfg.MaxConns))
	f.checks["TLS"].SetChecked(cfg.TLS)
	f.entries["CertPath"].SetText(cfg.CertPath)
	f.entries["KeyPath"].SetText(cfg.KeyPath)
	f.checks["MetricsEnabled"].SetChecked(cfg.MetricsEnabled)
	f.entries["MetricsPath"].SetText(cfg.MetricsPath)
	f.entries["LogFile"].SetText(cfg.LogFile)
	f.checks["AccessLog"].SetChecked(cfg.AccessLog)
	f.entries["MemLimitGB"].SetText(strconv.Itoa(cfg.MemLimitGB))
	f.entries["SwapGB"].SetText(strconv.Itoa(cfg.SwapGB))
	f.entries["NumaNode"].SetText(strconv.Itoa(cfg.NumaNode))
	f.entries["Priority"].SetText(cfg.Priority)
	f.entries["AffinityMask"].SetText(cfg.AffinityMask)
	f.entries["CacheDir"].SetText(cfg.CacheDir)
	f.entries["CacheSizeMB"].SetText(strconv.Itoa(cfg.CacheSizeMB))
	f.entries["EvictPolicy"].SetText(cfg.EvictPolicy)
	f.checks["Prefetch"].SetChecked(cfg.Prefetch)
}

// CollectConfig 从控件收集当前配置（主线程轻量读）。
// 第二个返回值是中文解析错误列表（空串按默认值处理）。
func (f *NinferForm) CollectConfig() (config.Config, []string) {
	var cfg config.Config
	var errs []string

	cfg.Host = f.entries["Host"].Text
	cfg.Port, errs = collectInt(f.entries["Port"], "监听端口", cfg.Port, errs)
	cfg.LogLevel = f.entries["LogLevel"].Text
	cfg.Workers, errs = collectInt(f.entries["Workers"], "并发 worker 数", cfg.Workers, errs)
	cfg.ModelName = f.entries["ModelName"].Text
	cfg.ModelPath = f.entries["ModelPath"].Text
	cfg.CtxSize, errs = collectInt(f.entries["CtxSize"], "上下文长度", cfg.CtxSize, errs)
	cfg.Threads, errs = collectInt(f.entries["Threads"], "推理线程", cfg.Threads, errs)
	cfg.Quant = f.entries["Quant"].Text
	cfg.Temperature, errs = collectFloat(f.entries["Temperature"], "采样温度", cfg.Temperature, errs)
	cfg.TopP, errs = collectFloat(f.entries["TopP"], "top-p", cfg.TopP, errs)
	cfg.TopK, errs = collectInt(f.entries["TopK"], "top-k", cfg.TopK, errs)
	cfg.RepeatPenalty, errs = collectFloat(f.entries["RepeatPenalty"], "重复惩罚", cfg.RepeatPenalty, errs)
	cfg.MaxTokens, errs = collectInt(f.entries["MaxTokens"], "最大生成 token", cfg.MaxTokens, errs)
	cfg.APIKey = f.entries["APIKey"].Text
	cfg.CORSOrigins = f.entries["CORSOrigins"].Text
	cfg.TimeoutSec, errs = collectInt(f.entries["TimeoutSec"], "请求超时", cfg.TimeoutSec, errs)
	cfg.MaxConns, errs = collectInt(f.entries["MaxConns"], "最大并发连接", cfg.MaxConns, errs)
	cfg.TLS = f.checks["TLS"].Checked
	cfg.CertPath = f.entries["CertPath"].Text
	cfg.KeyPath = f.entries["KeyPath"].Text
	cfg.MetricsEnabled = f.checks["MetricsEnabled"].Checked
	cfg.MetricsPath = f.entries["MetricsPath"].Text
	cfg.LogFile = f.entries["LogFile"].Text
	cfg.AccessLog = f.checks["AccessLog"].Checked
	cfg.MemLimitGB, errs = collectInt(f.entries["MemLimitGB"], "内存上限", cfg.MemLimitGB, errs)
	cfg.SwapGB, errs = collectInt(f.entries["SwapGB"], "swap 上限", cfg.SwapGB, errs)
	cfg.NumaNode, errs = collectInt(f.entries["NumaNode"], "NUMA 节点", cfg.NumaNode, errs)
	cfg.Priority = f.entries["Priority"].Text
	cfg.AffinityMask = f.entries["AffinityMask"].Text
	cfg.CacheDir = f.entries["CacheDir"].Text
	cfg.CacheSizeMB, errs = collectInt(f.entries["CacheSizeMB"], "缓存大小", cfg.CacheSizeMB, errs)
	cfg.EvictPolicy = f.entries["EvictPolicy"].Text
	cfg.Prefetch = f.checks["Prefetch"].Checked
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
		fmt.Fprintf(&b, "✗ %s", localizeValidation(e))
	}
	f.errLabel.SetText(b.String())
}

// localizeValidation 把错误码串映射为中文；未知错误码原样保留。
func localizeValidation(code string) string {
	for prefix, zh := range validationZh {
		if !strings.HasPrefix(code, prefix) {
			continue
		}
		rest := strings.TrimPrefix(code, prefix)
		if strings.Contains(zh, "%s") {
			return fmt.Sprintf(zh, rest)
		}
		return zh
	}
	return code
}
