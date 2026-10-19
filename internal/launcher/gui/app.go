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
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"ninferrmanager/internal/launcher/config"
	"ninferrmanager/internal/launcher/process"
)

// windowTitle 窗口标题（计划 T6）。
const windowTitle = "NInfer 启动器"

// windowSize 初始窗口尺寸；计划要求最小 900x640，此处给足余量。
var windowSize = fyne.NewSize(1100, 720)

// logMinHeight 日志区最小高度（计划：日志区高 200）。
const logMinHeight = 200

// statusInterval 状态刷新周期（计划：启动后 1s 周期刷新）。
const statusInterval = time.Second

// App 组装窗口、表单、状态面板、日志视图与进程 Manager（T6 聚合根）。
type App struct {
	fyneApp   fyne.App
	window    fyne.Window
	form      *Form
	status    *StatusPanel
	logv      *LogView
	msgLbl    *widget.Label  // 保存/加载操作反馈行（与校验行分离）
	pickerBtn *widget.Button // 选择 ninfer.exe 文件对话框入口
	manager   *process.Manager
	ring      *process.RingLog

	mu        sync.Mutex
	startedAt time.Time // 最近一次进入 Running 的时间（doStart 写，tick 读）
	lastLog   string    // 最近一条 worker 日志行（OnLine 后台线程写）
	valSeq    int64     // 校验代际，丢弃过期的后台校验结果

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

	a.form = NewForm(window)
	a.logv = NewLogView(a.ring)
	a.msgLbl.Wrapping = fyne.TextWrapWord

	// 按钮回调接线（NewForm 不预接 OnTapped，由 App 统一注入）。
	a.form.onChange = a.onFormChange
	a.form.saveBtn.OnTapped = a.onSave
	a.form.loadBtn.OnTapped = a.onLoad
	a.pickerBtn = widget.NewButton("选择 ninfer.exe…", a.pickWorker)

	// 状态面板：启动/停止回调注入（主线程回调，立即返回）。
	a.status = NewStatusPanel(a.onStart, a.onStop)

	// 日志接线：worker OnLine → 环形缓冲（必须在 Start 之前完成）。
	a.wireLog()

	// 布局：顶部（状态行 + 校验行 + 操作行）/ 两栏主体——左侧 8 组参数 tab、右侧日志区。
	logArea := a.logv.Content() // *container.Scroll，可 SetMinSize
	logArea.SetMinSize(fyne.NewSize(0, logMinHeight))
	toolbar := container.NewHBox(a.pickerBtn, a.form.saveBtn, a.form.loadBtn, layout.NewSpacer(), a.msgLbl)
	top := container.NewVBox(a.status.Row(), a.form.errLabel, toolbar)
	split := container.NewHSplit(a.form.tabs, logArea)
	split.Offset = 0.62 // 左侧表单占多数，分割条可拖动调整
	window.SetContent(container.NewBorder(top, nil, nil, nil, split))
	window.Resize(windowSize)
	window.SetOnClosed(a.onClosed)

	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.tickDone = make(chan struct{})

	// 初始化后台派发：worker 路径选择 + 初始配置加载 + 1s 状态刷新。
	go a.initWorkerPath()
	go a.doLoadInitial()
	go a.tick()

	return a
}

// initWorkerPath 决定 worker 可执行文件路径（后台 goroutine，NewApp 时派发）：
// 存档有效 → 直接采用并提示；否则回到主线程弹出文件选择框，由用户指定，
// 选择结果经 SaveWorkerPathDefault 持久化（下次启动跳过选择框）。
func (a *App) initWorkerPath() {
	saved, err := config.LoadWorkerPath()
	switch {
	case err != nil:
		log.Printf("读取 worker 路径存档失败: %v，回退默认路径", err)
	case saved != "" && fileExists(saved):
		a.manager.SetPath(saved)
		fyne.Do(func() {
			a.msgLbl.SetText("✓ 使用已保存的 ninfer.exe：" + saved)
		})
		return
	case saved != "":
		log.Printf("已保存的 worker 路径不存在: %s，重新选择", saved)
	}
	fyne.Do(a.pickWorker)
}

// fileExists 路径存在且不是目录。
func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// pickWorker 弹出文件选择框（必须主线程调用：按钮回调或 initWorkerPath 经 fyne.Do），
// 仅过滤 .exe。用户确认后进入后台 goroutine 校验、持久化并注入 Manager。
func (a *App) pickWorker() {
	d := dialog.NewFileOpen(a.onWorkerPicked, a.window)
	d.SetFilter(storage.NewExtensionFileFilter([]string{".exe"}))
	// 打开时定位到当前生效路径的目录，减少再次翻找。
	if p := a.manager.Path(); p != "" && p != process.DefaultWorkerPath {
		if lister, err := storage.ListerForURI(storage.NewFileURI(filepath.Dir(p))); err == nil {
			d.SetLocation(lister)
		}
	}
	d.Show()
}

// onWorkerPicked 文件选择回调（主线程，来自 dialog）：立即关闭 reader 后
// 将校验与持久化派发到后台 goroutine，避免回调内做文件 IO。
func (a *App) onWorkerPicked(reader fyne.URIReadCloser, err error) {
	if err != nil {
		log.Printf("文件选择失败: %v", err)
		return
	}
	if reader == nil {
		a.msgLbl.SetText("未选择 ninfer.exe，当前使用默认路径：" + a.manager.Path())
		return
	}
	path := reader.URI().Path()
	if closeErr := reader.Close(); closeErr != nil {
		log.Printf("关闭文件选择结果失败: %v", closeErr)
	}
	go a.applyWorkerChoice(path)
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
		a.msgLbl.SetText("✓ ninfer.exe 路径已保存，下次启动直接使用：" + path)
	})
}

// Run 启动 Fyne 事件循环（必须在主 goroutine 调用，阻塞直至应用退出）。
// Fyne 不会自动显示窗口：必须先显式 Show，否则主循环空转、进程无窗口。
func (a *App) Run() {
	a.window.Show()
	a.fyneApp.Run()
	a.cancel()
	<-a.tickDone
}

// doLoadInitial 首次载入配置（goroutine，NewApp 时派发）。
// 失败回退 DefaultConfig，错误经 log.Printf 留痕。
func (a *App) doLoadInitial() {
	path, err := config.LoadConfigPath()
	if err != nil {
		log.Printf("解析默认配置路径失败: %v，使用默认配置", err)
		a.applyConfig(config.DefaultConfig(), "未找到配置文件，已载入默认值")
		return
	}
	cfg, err := config.LoadConfig(path)
	if err != nil {
		log.Printf("加载配置 %s 失败: %v，使用默认配置", path, err)
		a.applyConfig(config.DefaultConfig(), "未找到配置文件，已载入默认值")
		return
	}
	a.applyConfig(cfg, "✓ 配置已加载")
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

// onFormChange 字段变化回调（主线程，来自 Form.onChange）：主线程收集表单值，
// 后台校验后立即返回（避免 goroutine 内 fyne.Do 收集与写回之间的竞态）。
func (a *App) onFormChange() {
	seq := atomic.AddInt64(&a.valSeq, 1)
	cfg, _ := a.form.CollectConfig() // 主线程读控件合法；解析错误由 Validate 覆盖
	go a.doValidate(seq, cfg)
}

// doValidate 后台校验配置（goroutine）；代际过期则丢弃结果。
func (a *App) doValidate(seq int64, cfg config.Config) {
	cfg = config.Normalize(cfg)
	errs := config.Validate(cfg)
	if atomic.LoadInt64(&a.valSeq) != seq {
		return
	}
	fyne.Do(func() { a.form.SetValidationErrors(errs) })
}

// onStart 启动按钮回调（主线程）：先校验再禁用按钮，Start 入 goroutine。
func (a *App) onStart() {
	if a.manager.State() != process.Stopped {
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

// doStart 后台启动 worker（goroutine）：规范化 → 全量校验 → Start。
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

// onStop 停止按钮回调（主线程）：仅 Running 可停，Stop 入 goroutine。
func (a *App) onStop() {
	if a.manager.State() != process.Running {
		return
	}
	go a.manager.Stop()
}

// onSave 保存按钮回调（主线程，Form.saveBtn）：主线程收集配置，入 goroutine 写盘。
func (a *App) onSave() {
	cfg, errs := a.form.CollectConfig()
	go a.doSave(cfg, errs)
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
	savedPath, err := config.SaveConfigDefault(cfg)
	if err != nil {
		log.Printf("保存配置失败: %v", err)
		fyne.Do(func() { a.msgLbl.SetText("保存失败：" + err.Error()) })
		return
	}
	fyne.Do(func() { a.msgLbl.SetText("✓ 配置已保存至 " + savedPath) })
}

// onLoad 加载按钮回调（主线程，Form.loadBtn）：读盘入 goroutine，经 fyne.Do 回填。
func (a *App) onLoad() {
	go func() {
		path, err := config.LoadConfigPath()
		if err != nil {
			log.Printf("解析默认配置路径失败: %v", err)
			fyne.Do(func() { a.msgLbl.SetText("加载失败：无法确定配置文件路径") })
			return
		}
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
