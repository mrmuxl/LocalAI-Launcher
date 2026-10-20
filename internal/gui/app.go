// app.go 是 GUI 聚合根：窗口布局、启动/停止/保存/加载/校验的线程编排、1s 状态刷新。
//
// 线程模型（对齐 AGENTS.md）：
//   - 事件回调（Tapped/OnChanged/OnClosed）运行在主线程，必须立即返回；
//   - 启动/停止/保存/加载/校验等耗时逻辑放入后台 goroutine；
//   - 一切 UI 更新经 fyne.Do 调度回主线程；
//   - goroutine 内的错误一律 log.Printf，不得静默丢弃。
package gui

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"llmlauncher/internal/argv"
	"llmlauncher/internal/config"
	"llmlauncher/internal/process"
)

// windowTitle 窗口标题（计划 T6）。
const windowTitle = "NInfer 启动器"

// windowSize 初始窗口尺寸；计划要求最小 900x640，此处给足余量。
var windowSize = fyne.NewSize(1100, 720)

// statusInterval 状态刷新周期（计划：启动后 1s 周期刷新）。
const statusInterval = time.Second

// defaultLlamaWorkerPath llama.cpp worker 的默认可执行文件路径（相对工作目录）。
const defaultLlamaWorkerPath = "llama-kvmem-server.exe"

// App 组装窗口、双引擎表单、状态面板、日志视图与进程 Manager（T6 聚合根）。
type App struct {
	fyneApp   fyne.App
	window    fyne.Window
	form      *NinferForm // NInfer 引擎表单（tab 0）
	llamaForm *LlamaForm  // llama.cpp 引擎表单（tab 1）
	tabs      *container.AppTabs
	status    *StatusPanel
	logv      *LogView
	msgLbl    *widget.Label // 保存/加载操作反馈行（与校验行分离）
	manager   *process.Manager
	ring      *process.RingLog

	llamaWorkerPath string // llama.cpp worker 可执行文件路径（内存，启动用）

	mu           sync.Mutex
	startedAt    time.Time // 最近一次进入 Running 的时间（doStart 写，tick 读）
	lastLog      string    // 最近一条 worker 日志行（OnLine 后台线程写）
	valSeqNinfer int64     // NInfer 表单校验代际
	valSeqLlama  int64     // llama 表单校验代际

	ctx      context.Context
	cancel   context.CancelFunc
	tickDone chan struct{}
}

// NewApp 构建 App 与窗口（必须在主 goroutine 调用：app.New 约束）。
// worker 路径不在此决定：由 initWorkerPath 在后台读取存档，
// 无存档时经文件选择框由用户指定（选择结果持久化，下次启动直接使用）。
func NewApp() *App {
	fyneApp := app.New()
	window := fyneApp.NewWindow(windowTitle)

	a := &App{
		fyneApp: fyneApp,
		window:  window,
		ring:    process.NewRingLog(0), // 0 → process.DefaultRingSize
		manager: process.New(""),
		msgLbl:  widget.NewLabel(""),
	}

	a.form = NewNinferForm(window)
	a.llamaForm = NewLlamaForm(window)
	a.logv = NewLogView(a.ring)
	a.msgLbl.Wrapping = fyne.TextWrapWord

	// 按钮回调接线（表单不预接 OnTapped，由 App 统一注入）。
	a.form.onChange = a.onFormChange
	a.form.saveBtn.OnTapped = a.onSave
	a.form.loadBtn.OnTapped = a.onLoad
	a.form.onOpenWorker = a.pickWorker
	a.form.onWorkerPathSet = a.onWorkerPathSubmitted
	a.form.onOpenModel = a.pickModel
	a.llamaForm.onChange = a.onFormChange
	a.llamaForm.onOpenWorker = a.pickLlamaWorker
	a.llamaForm.onWorkerPathSet = a.onLlamaWorkerPathSubmitted

	// 状态面板：启动/停止回调注入（主线程回调，立即返回）。
	a.status = NewStatusPanel(a.onStart, a.onStop)

	// 日志接线：worker OnLine → 环形缓冲（必须在 Start 之前完成）。
	a.wireLog()

	a.buildLayout(window)
	window.Resize(windowSize)
	window.SetOnClosed(a.onClosed)

	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.tickDone = make(chan struct{})

	// 初始化后台派发：worker 路径确定后加载配置（配置文件放推理程序目录）+ 1s 状态刷新。
	go a.initWorkerPath()
	go a.tick()

	return a
}

// initWorkerPath 决定两个引擎的 worker 可执行文件路径（后台 goroutine，NewApp 时派发）：
// 各读各的存档字段——有效则直接采用并回填；否则静默回退默认路径，不自动弹窗，
// 由用户点各 tab 的「打开」按钮时再弹选择框。
// 路径确定后（无论哪条分支）调用 doLoadInitial：配置文件放推理程序目录，
// 必须先定路径才能定位配置。
func (a *App) initWorkerPath() {
	saved, err := config.LoadWorkerPath()
	switch {
	case err != nil:
		log.Printf("读取 worker 路径存档失败: %v，回退默认路径", err)
	case saved != "" && fileExists(saved):
		a.manager.SetPath(saved)
		fyne.Do(func() {
			a.form.SetWorkerPath(saved)
			a.msgLbl.SetText("✓ 使用已保存的 ninfer.exe：" + saved)
		})
	default:
		if saved != "" {
			log.Printf("已保存的 worker 路径不存在: %s，回退默认路径", saved)
		}
		fyne.Do(func() {
			a.msgLbl.SetText("未指定推理程序，使用默认路径：" + a.manager.Path() + "；点「打开」选择 ninfer.exe")
		})
	}

	lsaved, lerr := config.LoadLlamaWorkerPath()
	switch {
	case lerr != nil:
		log.Printf("读取 llama worker 路径存档失败: %v，回退默认路径", lerr)
	case lsaved != "" && fileExists(lsaved):
		a.llamaWorkerPath = lsaved
		fyne.Do(func() { a.llamaForm.SetWorkerPath(lsaved) })
	case lsaved != "":
		log.Printf("已保存的 llama worker 路径不存在: %s，回退默认路径", lsaved)
	}

	a.doLoadInitial()
}

// fileExists 路径存在且不是目录。
func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// pickWorker 弹出 ninfer worker 文件选择框（主线程，按钮回调）。
func (a *App) pickWorker() { a.pickWorkerDialog(a.applyWorkerChoice) }

// pickLlamaWorker 弹出 llama.cpp worker 文件选择框（主线程，按钮回调）。
func (a *App) pickLlamaWorker() { a.pickWorkerDialog(a.applyLlamaWorkerChoice) }

// pickModel 弹出模型文件选择框（主线程，按钮回调）：过滤 .ninfer，
// 选择结果回填模型路径输入框（随配置保存，不单独持久化）。
func (a *App) pickModel() {
	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			log.Printf("模型文件选择失败: %v", err)
			return
		}
		if reader == nil {
			a.msgLbl.SetText("未选择模型文件，保持当前路径")
			return
		}
		path := reader.URI().Path()
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("关闭模型文件选择结果失败: %v", closeErr)
		}
		fyne.Do(func() { a.form.SetModelPath(path) })
	}, a.window)
	d.SetFilter(storage.NewExtensionFileFilter([]string{".ninfer"}))
	d.Show()
}

// pickWorkerDialog 弹出文件选择框（必须主线程调用），仅过滤 .exe，
// 用户确认后进入后台 goroutine 执行 apply（校验、持久化并注入）。
func (a *App) pickWorkerDialog(apply func(path string)) {
	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		// dialog 回调运行在主线程：立即关闭 reader，IO 派发后台。
		if err != nil {
			log.Printf("文件选择失败: %v", err)
			return
		}
		if reader == nil {
			a.msgLbl.SetText("未选择文件，保持当前路径")
			return
		}
		path := reader.URI().Path()
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("关闭文件选择结果失败: %v", closeErr)
		}
		go apply(path)
	}, a.window)
	d.SetFilter(storage.NewExtensionFileFilter([]string{".exe"}))
	// 打开时定位到当前生效路径的目录，减少再次翻找。
	if p := a.manager.Path(); p != "" && p != process.DefaultWorkerPath {
		if lister, err := storage.ListerForURI(storage.NewFileURI(filepath.Dir(p))); err == nil {
			d.SetLocation(lister)
		}
	}
	d.Show()
}

// applyWorkerChoice 校验、持久化并注入用户选择的 worker 路径（后台 goroutine，
// 由 onWorkerPicked 派发）。持久化失败仍注入内存路径，仅日志留痕提示未保存。
func (a *App) applyWorkerChoice(path string) {
	if !fileExists(path) {
		log.Printf("选择的 worker 路径不存在: %s", path)
		fyne.Do(func() {
			a.msgLbl.SetText("✗ 所选文件不存在，请重新选择")
		})
		return
	}

	if savedPath, err := config.SaveWorkerPathDefault(path); err != nil {
		log.Printf("保存 worker 路径失败: %v", err)
	} else {
		log.Printf("worker 路径已保存: %s", savedPath)
	}

	a.manager.SetPath(path)
	fyne.Do(func() {
		a.form.SetWorkerPath(path)
		a.msgLbl.SetText("✓ ninfer.exe 路径已保存，下次启动直接使用：" + path)
	})
}

// onWorkerPathSubmitted 推理程序输入框回车回调（主线程，来自 NinferForm）：
// 立即返回，校验/持久化/注入派发到后台 goroutine。
func (a *App) onWorkerPathSubmitted(path string) {
	go a.applyWorkerChoice(path)
}

// onLlamaWorkerPathSubmitted llama 推理程序输入框回车回调（主线程，来自 LlamaForm）。
func (a *App) onLlamaWorkerPathSubmitted(path string) {
	go a.applyLlamaWorkerChoice(path)
}

// applyLlamaWorkerChoice 校验、持久化并注入用户选择的 llama worker 路径
// （后台 goroutine）。持久化失败仍注入内存路径，仅日志留痕。
func (a *App) applyLlamaWorkerChoice(path string) {
	if !fileExists(path) {
		log.Printf("选择的 llama worker 路径不存在: %s", path)
		fyne.Do(func() { a.msgLbl.SetText("✗ 所选文件不存在，请重新选择") })
		return
	}
	if savedPath, err := config.SaveLlamaWorkerPath(path); err != nil {
		log.Printf("保存 llama worker 路径失败: %v", err)
	} else {
		log.Printf("llama worker 路径已保存: %s", savedPath)
	}
	a.llamaWorkerPath = path
	fyne.Do(func() {
		a.llamaForm.SetWorkerPath(path)
		a.msgLbl.SetText("✓ llama worker 路径已保存，下次启动直接使用：" + path)
	})
}

// llamaWorkerPathOrDef 返回当前 llama worker 路径（空则默认相对路径）。
func (a *App) llamaWorkerPathOrDef() string {
	if a.llamaWorkerPath != "" {
		return a.llamaWorkerPath
	}
	return defaultLlamaWorkerPath
}

// Run 启动 Fyne 事件循环（必须在主 goroutine 调用，阻塞直至应用退出）。
// Fyne 不会自动显示窗口：必须先显式 Show，否则主循环空转、进程无窗口。
func (a *App) Run() {
	a.window.Show()
	a.fyneApp.Run()
	a.cancel()
	<-a.tickDone
}

// doLoadInitial 首次载入两个引擎的配置（initWorkerPath 末尾调用，路径已确定）。
// 各引擎配置放各自推理程序所在目录；失败回退默认值，错误经 log.Printf 留痕。
func (a *App) doLoadInitial() {
	path := config.ConfigPathFor(a.manager.Path(), "config.json")
	cfg, err := config.LoadConfig(path)
	if err != nil {
		log.Printf("加载配置 %s 失败: %v，使用默认配置", path, err)
		a.applyConfig(config.DefaultConfig(), "未找到配置文件，已载入默认值")
	} else {
		a.applyConfig(cfg, "✓ 配置已加载")
	}

	lpath := config.ConfigPathFor(a.llamaWorkerPathOrDef(), "llama.json")
	lcfg, lerr := config.LoadLlamaConfig(lpath)
	if lerr != nil {
		log.Printf("加载 llama 配置 %s 失败: %v，使用默认配置", lpath, lerr)
		a.applyLlamaConfig(config.DefaultLlamaConfig(), "llama 配置未找到，已载入默认值")
		return
	}
	a.applyLlamaConfig(lcfg, "✓ llama 配置已加载")
}

// applyConfig 规范化并回填表单 + 校验 + 反馈文案（goroutine 内调用）。
// Validate 含文件存在性检查（fs IO），只可在 goroutine 中执行。
func (a *App) applyConfig(cfg config.Config, msg string) {
	cfg = config.Normalize(cfg)
	errs := config.Validate(cfg)
	fyne.Do(func() {
		a.form.SetFieldsFrom(cfg)
		a.form.SetValidationErrors(errs)
		a.msgLbl.SetText(msg)
	})
}

// applyLlamaConfig 回填 llama 表单 + 校验 + 反馈文案（goroutine 内调用）。
// ValidateLlama 含模型文件存在性检查（fs IO），只可在 goroutine 中执行。
func (a *App) applyLlamaConfig(cfg config.LlamaConfig, msg string) {
	errs := config.ValidateLlama(cfg)
	fyne.Do(func() {
		a.llamaForm.SetFieldsFrom(cfg)
		a.llamaForm.SetValidationErrors(errs)
		a.msgLbl.SetText(msg)
	})
}

// tick 每 1s 经 fyne.Do 刷新状态行与日志区（后台 goroutine，ctx 取消后退出）。
func (a *App) tick() {
	defer close(a.tickDone)
	t := time.NewTicker(statusInterval)
	defer t.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-t.C:
			st := a.manager.State()
			pid := a.manager.Pid()
			a.mu.Lock()
			startedAt := a.startedAt
			lastLog := a.lastLog
			a.mu.Unlock()
			fyne.Do(func() {
				a.status.ApplyState(st, pid, startedAt, lastLog)
				a.logv.Refresh()
			})
		}
	}
}

// onFormChange 字段变化回调（主线程，来自两个表单的 onChange）：按当前选中
// tab 收集对应引擎的表单值，后台校验后立即返回（避免 goroutine 内
// fyne.Do 收集与写回之间的竞态）。
func (a *App) onFormChange() {
	if a.tabs != nil && a.tabs.SelectedIndex() == 1 {
		seq := atomic.AddInt64(&a.valSeqLlama, 1)
		cfg, _ := a.llamaForm.CollectConfig() // 主线程读控件合法
		go a.doValidateLlama(seq, cfg)
		return
	}
	seq := atomic.AddInt64(&a.valSeqNinfer, 1)
	cfg, _ := a.form.CollectConfig() // 主线程读控件合法；解析错误由 Validate 覆盖
	go a.doValidateNinfer(seq, cfg)
}

// doValidateNinfer 后台校验 ninfer 配置（goroutine）；代际过期则丢弃结果。
func (a *App) doValidateNinfer(seq int64, cfg config.Config) {
	cfg = config.Normalize(cfg)
	errs := config.Validate(cfg)
	if atomic.LoadInt64(&a.valSeqNinfer) != seq {
		return
	}
	fyne.Do(func() { a.form.SetValidationErrors(errs) })
}

// doValidateLlama 后台校验 llama 配置（goroutine）；代际过期则丢弃结果。
func (a *App) doValidateLlama(seq int64, cfg config.LlamaConfig) {
	errs := config.ValidateLlama(cfg)
	if atomic.LoadInt64(&a.valSeqLlama) != seq {
		return
	}
	fyne.Do(func() { a.llamaForm.SetValidationErrors(errs) })
}

// onStart 启动按钮回调（主线程）：按当前选中 tab 选择引擎，
// 先校验再禁用按钮，Start 入 goroutine。
func (a *App) onStart() {
	if a.manager.State() != process.Stopped {
		return
	}
	if a.tabs != nil && a.tabs.SelectedIndex() == 1 {
		cfg, errs := a.llamaForm.CollectConfig()
		if len(errs) > 0 {
			a.llamaForm.SetValidationErrors(errs)
			log.Printf("拒绝启动：llama 配置解析未通过（%d 项）", len(errs))
			return
		}
		a.status.startBtn.Disable()
		go a.doStartLlama(cfg)
		return
	}
	cfg, errs := a.form.CollectConfig()
	if len(errs) > 0 {
		a.form.SetValidationErrors(errs)
		log.Printf("拒绝启动：配置校验未通过（%d 项）", len(errs))
		return
	}
	a.status.startBtn.Disable()
	go a.doStart(cfg)
}

// doStartLlama 后台启动 llama worker（goroutine）：全量校验 → 输出命令行 → StartPath。
func (a *App) doStartLlama(cfg config.LlamaConfig) {
	if errs := config.ValidateLlama(cfg); len(errs) > 0 {
		log.Printf("启动前校验失败: %v", errs)
		fyne.Do(func() {
			a.llamaForm.SetValidationErrors(errs)
			a.status.startBtn.Enable()
		})
		return
	}
	exe := a.llamaWorkerPathOrDef()
	args := argv.BuildLlamaArgs(cfg)
	a.announceCmdline(exe, args)
	if err := a.manager.StartPath(exe, args); err != nil {
		log.Printf("启动 worker 失败: %v", err)
		fyne.Do(func() { a.status.startBtn.Enable() })
		return
	}
	a.mu.Lock()
	a.startedAt = time.Now()
	a.mu.Unlock()
	log.Printf("worker 已启动 (PID=%d)", a.manager.Pid())
}

// doStart 后台启动 worker（goroutine）：规范化 → 全量校验 → 输出命令行 → Start。
func (a *App) doStart(cfg config.Config) {
	cfg = config.Normalize(cfg)
	if errs := config.Validate(cfg); len(errs) > 0 {
		log.Printf("启动前校验失败: %v", errs)
		fyne.Do(func() {
			a.form.SetValidationErrors(errs)
			a.status.startBtn.Enable()
		})
		return
	}
	exe := a.manager.Path()
	args := argv.BuildArgs(cfg)
	// GPU 选择经环境变量传递（对齐 start-backend.ps1：CUDA_VISIBLE_DEVICES + 固定 PCI_BUS_ID 排序）。
	a.manager.SetEnv("CUDA_VISIBLE_DEVICES", cfg.Gpu)
	a.manager.SetEnv("CUDA_DEVICE_ORDER", "PCI_BUS_ID")
	a.announceCmdline(exe, args)
	if err := a.manager.Start(cfg); err != nil {
		log.Printf("启动 worker 失败: %v", err)
		fyne.Do(func() { a.status.startBtn.Enable() })
		return
	}
	a.mu.Lock()
	a.startedAt = time.Now()
	a.mu.Unlock()
	log.Printf("worker 已启动 (PID=%d)", a.manager.Pid())
}

// announceCmdline 把即将执行的完整命令行写入日志面板与标准日志
// （goroutine 内调用；ring.Append 线程安全，UI 由 tick 周期刷新，此处立即刷一次）。
func (a *App) announceCmdline(exe string, args []string) {
	cmdline := exe + " " + strings.Join(args, " ")
	a.ring.Append("$ " + cmdline)
	log.Printf("启动命令行: %s", cmdline)
	fyne.Do(func() { a.logv.Refresh() })
}

// onStop 停止按钮回调（主线程）：仅 Running 可停，Stop 入 goroutine。
func (a *App) onStop() {
	if a.manager.State() != process.Running {
		return
	}
	go a.manager.Stop()
}

// onSave 保存按钮回调（主线程）：按当前选中 tab 收集配置，入 goroutine 写盘。
func (a *App) onSave() {
	if a.tabs != nil && a.tabs.SelectedIndex() == 1 {
		cfg, errs := a.llamaForm.CollectConfig()
		go a.doSaveLlama(cfg, errs)
		return
	}
	cfg, errs := a.form.CollectConfig()
	go a.doSave(cfg, errs)
}

// doSaveLlama 后台写 llama 配置（goroutine）：先校验，失败则回显。
func (a *App) doSaveLlama(cfg config.LlamaConfig, errs []string) {
	if len(errs) == 0 {
		errs = config.ValidateLlama(cfg)
	}
	if len(errs) > 0 {
		log.Printf("保存 llama 配置失败（校验未通过）: %v", errs)
		fyne.Do(func() {
			a.llamaForm.SetValidationErrors(errs)
			a.msgLbl.SetText("保存失败：请先修正校验错误")
		})
		return
	}
	path := config.ConfigPathFor(a.llamaWorkerPathOrDef(), "llama.json")
	if err := config.SaveLlamaConfig(cfg, path); err != nil {
		log.Printf("保存 llama 配置失败: %v", err)
		fyne.Do(func() { a.msgLbl.SetText("保存失败：" + err.Error()) })
		return
	}
	fyne.Do(func() { a.msgLbl.SetText("✓ llama 配置已保存至 " + path) })
}

// doSave 后台写配置（goroutine）：先校验（fs），失败则回显。
func (a *App) doSave(cfg config.Config, errs []string) {
	if len(errs) == 0 {
		cfg = config.Normalize(cfg)
		errs = config.Validate(cfg)
	}
	if len(errs) > 0 {
		log.Printf("保存配置失败（校验未通过）: %v", errs)
		fyne.Do(func() {
			a.form.SetValidationErrors(errs)
			a.msgLbl.SetText("保存失败：请先修正校验错误")
		})
		return
	}
	path := config.ConfigPathFor(a.manager.Path(), "config.json")
	if err := config.SaveConfig(cfg, path); err != nil {
		log.Printf("保存配置失败: %v", err)
		fyne.Do(func() { a.msgLbl.SetText("保存失败：" + err.Error()) })
		return
	}
	fyne.Do(func() { a.msgLbl.SetText("✓ 配置已保存至 " + path) })
}

// onLoad 加载按钮回调（主线程）：按当前选中 tab 读对应引擎配置，
// 读盘入 goroutine，经 fyne.Do 回填。
func (a *App) onLoad() {
	if a.tabs != nil && a.tabs.SelectedIndex() == 1 {
		go func() {
			path := config.ConfigPathFor(a.llamaWorkerPathOrDef(), "llama.json")
			cfg, err := config.LoadLlamaConfig(path)
			if err != nil {
				log.Printf("加载 llama 配置 %s 失败: %v", path, err)
				fyne.Do(func() { a.msgLbl.SetText("加载失败：llama 配置文件不存在或已损坏，请检查") })
				return
			}
			a.applyLlamaConfig(cfg, "✓ llama 配置已加载")
		}()
		return
	}
	go func() {
		path := config.ConfigPathFor(a.manager.Path(), "config.json")
		cfg, err := config.LoadConfig(path)
		if err != nil {
			log.Printf("加载配置 %s 失败: %v", path, err)
			fyne.Do(func() { a.msgLbl.SetText("加载失败：配置文件中不存在或已损坏，请检查") })
			return
		}
		a.applyConfig(cfg, "✓ 配置已加载")
	}()
}

// onClosed 窗口关闭回调（主线程）：取消刷新循环，尽力停止 worker。
func (a *App) onClosed() {
	a.cancel()
	if a.manager.State() != process.Stopped {
		go func() {
			if err := a.manager.Stop(); err != nil {
				log.Printf("关闭时停止 worker 失败: %v", err)
			}
		}()
	}
}

// wireLog 把 Manager 的日志行接入环形缓冲（主线程，构造期调用）。
// OnLine 回调运行在 Manager 的后台读行 goroutine：只写 ring 与 lastLog（均线程安全），不碰 UI。
func (a *App) wireLog() {
	a.manager.OnLine = func(line string) {
		a.ring.Append(line)
		a.mu.Lock()
		a.lastLog = line
		a.mu.Unlock()
	}
}
