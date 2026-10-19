package process

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"ninferrmanager/internal/launcher/argv"
	"ninferrmanager/internal/launcher/config"
)

// State 描述 Manager 当前所处状态，对应计划 D4 状态机：
// stopped → starting → running → stopping → stopped。
type State int

// 状态枚举。
const (
	Stopped  State = iota // 未启动或已完全退出
	Starting              // Start 进行中（进程尚未确认存活）
	Running               // 进程运行中
	Stopping              // Stop 已发出，等待宽限期退出或强 kill
)

// String 返回状态的可读文本（用于日志，非 UI 文案）。
func (s State) String() string {
	switch s {
	case Stopped:
		return "stopped"
	case Starting:
		return "starting"
	case Running:
		return "running"
	case Stopping:
		return "stopping"
	default:
		return "unknown"
	}
}

// 哨兵错误。
var (
	// ErrAlreadyRunning 进程正在启动/运行/停止时再次 Start（并发 Start 只生效一次）。
	ErrAlreadyRunning = errors.New("process: 进程已在运行或停止中，拒绝重复 Start")
	// ErrNotStarted Wait 时从未成功 Start 过。
	ErrNotStarted = errors.New("process: 进程未启动，无法 Wait")
)

// DefaultWorkerPath 默认 worker 可执行文件路径（相对当前工作目录）。
// 实际部署路径由上层（GUI）在构造 Manager 时注入。
const DefaultWorkerPath = "ninfer.exe"

// GracePeriod 计划「线程与 UI 契约」约定的 Stop 宽限期：3 秒，超时后强 kill。
const GracePeriod = 3 * time.Second

// Manager 管理一个 worker 子进程：Start / Stop / Wait 全生命周期。
//
// 线程模型（对齐 AGENTS.md）：
//   - Start 派生的三个后台 goroutine（stdout 读行、stderr 读行、Wait 收尾）
//     全部在进程退出后自然结束，不泄漏；
//   - OnLine 回调在后台 goroutine 中调用，回调方自行保证线程安全；
//   - 本包所有方法均可并发调用：重复/并发 Start 只生效一次，Stop/Wait 幂等。
type Manager struct {
	cmdPath string        // worker 可执行文件路径（可注入，测试用 fake 命令）
	Grace   time.Duration // Stop 宽限期，超时后强 kill；默认 GracePeriod
	// OnLine 日志行回调（stdout/stderr 每行一次），必须在 Start 前设置。
	OnLine func(line string)

	mu      sync.Mutex
	state   State
	cmd     *exec.Cmd
	done    chan struct{} // 进程完全退出（管道读完 + Wait 返回）后关闭
	code    int           // 进程退出码
	waitErr error         // cmd.Wait 的错误（被 kill 时为 *exec.ExitError）
}

// New 创建 Manager；cmdPath 为空时使用 DefaultWorkerPath。
func New(cmdPath string) *Manager {
	if cmdPath == "" {
		cmdPath = DefaultWorkerPath
	}
	return &Manager{cmdPath: cmdPath, Grace: GracePeriod}
}

// SetPath 更新 worker 可执行文件路径（如用户通过文件选择框重新指定）。
// 仅在 Stopped 状态生效；p 为空时忽略，保持现有路径。
func (m *Manager) SetPath(p string) {
	if p == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.state == Stopped {
		m.cmdPath = p
	}
}

// Path 返回当前生效的 worker 可执行文件路径。
func (m *Manager) Path() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cmdPath
}

// State 返回当前状态。
func (m *Manager) State() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// Pid 返回当前 worker 的操作系统进程号；进程不存在或尚未启动时返回 0。
func (m *Manager) Pid() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cmd != nil && m.cmd.Process != nil {
		return m.cmd.Process.Pid
	}
	return 0
}

// Start 用 argv.BuildArgs(cfg) 构造参数并启动 worker 进程。
// 并发/重复 Start 只生效一次，其余调用返回 ErrAlreadyRunning。
// 进程完全退出后状态复位为 Stopped，Manager 可再次 Start。
func (m *Manager) Start(cfg config.Config) error {
	args := argv.BuildArgs(cfg)
	cmd := exec.Command(m.cmdPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	m.mu.Lock()
	if m.state != Stopped {
		m.mu.Unlock()
		return ErrAlreadyRunning
	}
	m.state = Starting
	m.cmd = cmd
	m.code = 0
	m.waitErr = nil
	done := make(chan struct{})
	m.done = done
	cb := m.OnLine
	m.mu.Unlock()

	if err := cmd.Start(); err != nil {
		m.finishStartFailure(done)
		return err
	}

	m.mu.Lock()
	m.state = Running
	m.mu.Unlock()

	// 三个后台 goroutine：stdout 读行、stderr 读行、Wait 收尾。
	// 进程退出后三者全部自然结束，无泄漏。
	var wg sync.WaitGroup
	wg.Add(2)
	go m.readLines(&wg, stdout, cb)
	go m.readLines(&wg, stderr, cb)
	go func() {
		wg.Wait()
		waitErr := cmd.Wait()

		m.mu.Lock()
		m.state = Stopped
		m.cmd = nil
		m.waitErr = waitErr
		if cmd.ProcessState != nil {
			m.code = cmd.ProcessState.ExitCode()
		}
		m.mu.Unlock()

		close(done)
	}()
	return nil
}

// finishStartFailure 在 cmd.Start 失败时复位状态并关闭 done，
// 避免并发 Wait 永久阻塞。
func (m *Manager) finishStartFailure(done chan struct{}) {
	m.mu.Lock()
	m.state = Stopped
	m.cmd = nil
	m.mu.Unlock()
	close(done)
}

// readLines 逐行读取 r 直到 EOF，逐行调用 cb；EOF 后 defer wg.Done()
// 通知收尾 goroutine 可执行 cmd.Wait()（否则计数器永不归零，Stop 永久阻塞）。
// 进程退出或被 kill 时管道以 EOF 关闭，属预期路径，无错误需要上报。
func (m *Manager) readLines(wg *sync.WaitGroup, r io.Reader, cb func(line string)) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		if cb != nil {
			cb(scanner.Text())
		}
	}
}

// Stop 请求进程退出：先尝试 os.Interrupt 优雅信号（Windows 标准库不支持时
// 记录后直接进入宽限期），宽限期 Grace 内进程未退出则强 Kill。
// 幂等：未启动、已退出、重复 Stop 均返回 nil，不 panic。
func (m *Manager) Stop() error {
	m.mu.Lock()
	cmd := m.cmd
	done := m.done
	if m.state == Running {
		m.state = Stopping
	}
	m.mu.Unlock()

	if cmd == nil || cmd.Process == nil || done == nil {
		return nil
	}

	// 优雅信号；Windows 上 os.Process.Signal 不受支持，忽略错误继续走宽限期。
	if sigErr := cmd.Process.Signal(os.Interrupt); sigErr != nil {
		log.Printf("process: 发送 os.Interrupt 失败（Windows 预期内，转宽限期）: %v", sigErr)
	}

	select {
	case <-done:
		// 宽限期内自然退出（或已退出）
		return nil
	case <-time.After(m.Grace):
		if killErr := cmd.Process.Kill(); killErr != nil {
			log.Printf("process: 强 kill 失败: %v", killErr)
		}
		<-done
		return nil
	}
}

// Wait 阻塞至进程完全退出（含管道读完），返回退出码。
// 从未成功 Start 过返回 ErrNotStarted。
func (m *Manager) Wait() (int, error) {
	m.mu.Lock()
	done := m.done
	m.mu.Unlock()
	if done == nil {
		return 0, ErrNotStarted
	}
	<-done

	m.mu.Lock()
	defer m.mu.Unlock()
	return m.code, m.waitErr
}
