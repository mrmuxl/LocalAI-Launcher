package process

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"llmlauncher/internal/config"
)

// ---------- fake 命令构造 ----------
//
// Windows 环境无 Unix echo 语义可依赖，统一用 go build 在 t.TempDir()
// 编译临时小 exe 作为 fake worker（忽略全部命令行参数），禁止硬编码绝对路径。

const fakeSlowSrc = `package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("worker: line one")
	fmt.Println("worker: line two")
	time.Sleep(60 * time.Second)
}
`

const fakeExit7Src = `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("worker: about to exit 7")
	os.Exit(7)
}
`

const fakeEchoSrc = `package main

import "fmt"

func main() {
	fmt.Println("hello from fake worker")
}
`

// buildFake 编译一个 fake worker exe 到 t.TempDir() 并返回其路径。
func buildFake(t *testing.T, source string) string {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "fake.go")
	if err := os.WriteFile(src, []byte(source), 0o644); err != nil {
		t.Fatalf("写 fake 源码: %v", err)
	}
	exe := filepath.Join(dir, "fake")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	out, err := exec.Command("go", "build", "-o", exe, src).CombinedOutput()
	if err != nil {
		t.Fatalf("go build fake 失败: %v\n%s", err, out)
	}
	return exe
}

func defaultCfg() config.Config { return config.DefaultConfig() }

// lineCollector 线程安全地收集 OnLine 回调行。
type lineCollector struct {
	mu    sync.Mutex
	lines []string
}

func (c *lineCollector) add(line string) {
	c.mu.Lock()
	c.lines = append(c.lines, line)
	c.mu.Unlock()
}

func (c *lineCollector) snapshot() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.lines...)
}

// waitLines 阻塞直到收集到至少 n 行或超时，返回当前已收到的行。
func (c *lineCollector) waitLines(t *testing.T, n int, timeout time.Duration) []string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if got := c.snapshot(); len(got) >= n {
			return got
		}
		if time.Now().After(deadline) {
			return c.snapshot()
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// waitExitCode 执行 Wait 并断言错误语义：nil 或 *exec.ExitError（被杀/非零退出的
// 预期路径，对齐 exec.Cmd.Wait 惯例）均合法，其他错误判失败。返回退出码。
func waitExitCode(t *testing.T, m *Manager) int {
	t.Helper()
	code, err := m.Wait()
	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			t.Fatalf("Wait: %v", err)
		}
	}
	return code
}

// ---------- 生命周期 ----------

// TestManagerLifecycleAndRestart 验证完整生命周期：
// Start → 行回调 → Stop（宽限期后强 kill）→ Wait → 状态复位 → 再 Start 仍可收行。
// 「进程被杀后 Manager 可再 Start」是本计划验收硬条件。
func TestManagerLifecycleAndRestart(t *testing.T) {
	exe := buildFake(t, fakeSlowSrc)
	m := New(exe)
	m.Grace = 300 * time.Millisecond // 加速；默认 3s 宽限期由 TestManagerStopKillsAfterDefaultGrace 验证

	c := &lineCollector{}
	m.OnLine = c.add

	if err := m.Start(defaultCfg()); err != nil {
		t.Fatalf("首次 Start: %v", err)
	}
	if got := m.State(); got != Running {
		t.Fatalf("Start 后状态 = %v, want Running", got)
	}

	got := c.waitLines(t, 2, 5*time.Second)
	if len(got) < 2 || got[0] != "worker: line one" || got[1] != "worker: line two" {
		t.Fatalf("收到行 = %v, want [worker: line one, worker: line two]", got)
	}

	if err := m.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	code := waitExitCode(t, m)
	if code == 0 {
		t.Fatalf("被强 kill 的进程退出码 = 0，非零退出码才是被杀语义")
	}
	if got := m.State(); got != Stopped {
		t.Fatalf("Stop+Wait 后状态 = %v, want Stopped", got)
	}

	// 进程被杀后 Manager 可再 Start
	c2 := &lineCollector{}
	m.OnLine = c2.add
	if err := m.Start(defaultCfg()); err != nil {
		t.Fatalf("再次 Start（被杀后）: %v", err)
	}
	got2 := c2.waitLines(t, 2, 5*time.Second)
	if len(got2) < 2 {
		t.Fatalf("再次 Start 后收到行 = %v, want 至少 2 行", got2)
	}
	if err := m.Stop(); err != nil {
		t.Fatalf("再次 Stop: %v", err)
	}
	_ = waitExitCode(t, m)
}

// TestManagerStopKillsAfterDefaultGrace 验证默认宽限期为 3s：
// fake 进程会睡眠 60s，Stop 必须等满宽限期后强 kill（Stop 耗时 >= 2.5s）。
func TestManagerStopKillsAfterDefaultGrace(t *testing.T) {
	exe := buildFake(t, fakeSlowSrc)
	m := New(exe) // Grace 保持默认 3s

	c := &lineCollector{}
	m.OnLine = c.add
	if err := m.Start(defaultCfg()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got := c.waitLines(t, 1, 5*time.Second); len(got) < 1 {
		t.Fatalf("未收到行: %v", got)
	}

	begin := time.Now()
	if err := m.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	elapsed := time.Since(begin)
	if elapsed < 2500*time.Millisecond {
		t.Fatalf("Stop 耗时 %v < 2.5s：默认宽限期 3s 未生效", elapsed)
	}
	waitExitCode(t, m)
}

// ---------- 重复 Start ----------

// TestManagerDuplicateStartRejected 重复 Start 必须被拒绝且只生效一次。
func TestManagerDuplicateStartRejected(t *testing.T) {
	exe := buildFake(t, fakeSlowSrc)
	m := New(exe)
	m.Grace = 300 * time.Millisecond

	c := &lineCollector{}
	m.OnLine = c.add

	if err := m.Start(defaultCfg()); err != nil {
		t.Fatalf("首次 Start: %v", err)
	}
	defer func() {
		if err := m.Stop(); err != nil {
			t.Errorf("清理 Stop: %v", err)
		}
		_, _ = m.Wait()
	}()

	if err := m.Start(defaultCfg()); !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("重复 Start 错误 = %v, want ErrAlreadyRunning", err)
	}

	// 并发 Start：Manager 已在 Running，全部被拒（互斥语义）
	var wg sync.WaitGroup
	errs := make([]error, 3)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = m.Start(defaultCfg())
		}(i)
	}
	wg.Wait()
	ok, rejected := 0, 0
	for _, err := range errs {
		if err == nil {
			ok++
		} else if errors.Is(err, ErrAlreadyRunning) {
			rejected++
		} else {
			t.Fatalf("并发 Start 意外错误: %v", err)
		}
	}
	if ok != 0 || rejected != 3 {
		t.Fatalf("并发 Start 成功/拒绝 = %d/%d, want 0/3", ok, rejected)
	}
}

// TestManagerConcurrentStartOnlyOneWins 全新 Manager 上 4 个 goroutine 并发 Start，
// 恰好 1 个成功、3 个被拒（并发 Start 只生效一次）。
func TestManagerConcurrentStartOnlyOneWins(t *testing.T) {
	exe := buildFake(t, fakeSlowSrc)
	m := New(exe)
	m.Grace = 300 * time.Millisecond

	const n = 4
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = m.Start(defaultCfg())
		}(i)
	}
	wg.Wait()
	ok, rejected := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, ErrAlreadyRunning):
			rejected++
		default:
			t.Fatalf("并发 Start 意外错误: %v", err)
		}
	}
	if ok != 1 || rejected != n-1 {
		t.Fatalf("并发 Start 成功/拒绝 = %d/%d, want 1/%d", ok, rejected, n-1)
	}
	// 清理：Stop 幂等，无论哪个 goroutine 成功都能安全停止
	if err := m.Stop(); err != nil {
		t.Fatalf("清理 Stop: %v", err)
	}
	waitExitCode(t, m)
}

// ---------- Stop 幂等 ----------

// TestManagerStopIdempotent 重复 Stop 不 panic 不报错；未启动 Stop 亦安全。
func TestManagerStopIdempotent(t *testing.T) {
	// 未启动时 Stop
	m0 := New("nonexistent-fake")
	if err := m0.Stop(); err != nil {
		t.Fatalf("未启动 Stop 应返回 nil, got %v", err)
	}

	// 进程自然退出后连续多次 Stop
	exe := buildFake(t, fakeEchoSrc)
	m := New(exe)
	m.Grace = 300 * time.Millisecond
	if err := m.Start(defaultCfg()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	waitExitCode(t, m)
	for i := 0; i < 3; i++ {
		if err := m.Stop(); err != nil {
			t.Fatalf("第 %d 次重复 Stop 应返回 nil, got %v", i+1, err)
		}
	}

	// Stop 之后立即重复 Stop（进程运行中连 Stop）
	exe2 := buildFake(t, fakeSlowSrc)
	m2 := New(exe2)
	m2.Grace = 200 * time.Millisecond
	if err := m2.Start(defaultCfg()); err != nil {
		t.Fatalf("Start m2: %v", err)
	}
	if err := m2.Stop(); err != nil {
		t.Fatalf("Stop m2: %v", err)
	}
	if err := m2.Stop(); err != nil {
		t.Fatalf("重复 Stop m2 应返回 nil, got %v", err)
	}
	waitExitCode(t, m2)
}

// ---------- 退出码 ----------

// TestManagerWaitExitCode Wait 返回真实退出码（非 0 的自然退出）。
func TestManagerWaitExitCode(t *testing.T) {
	exe := buildFake(t, fakeExit7Src)
	m := New(exe)
	if err := m.Start(defaultCfg()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	code := waitExitCode(t, m)
	if code != 7 {
		t.Fatalf("退出码 = %d, want 7", code)
	}
}

// ---------- goroutine 泄漏 ----------

// TestManagerGoroutineNoLeak 进程完全退出后，Start 派生的后台 goroutine
// （stdout 读行 + stderr 读行 + Wait 收尾）必须全部结束。
func TestManagerGoroutineNoLeak(t *testing.T) {
	baseline := time.Now()
	time.Sleep(150 * time.Millisecond) // 让先例测试的收尾 goroutine 完全退出
	base := runtime.NumGoroutine()

	exe := buildFake(t, fakeSlowSrc)
	m := New(exe)
	m.Grace = 300 * time.Millisecond
	c := &lineCollector{}
	m.OnLine = c.add

	if err := m.Start(defaultCfg()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got := c.waitLines(t, 1, 5*time.Second); len(got) < 1 {
		t.Fatalf("未收到行: %v", got)
	}
	if err := m.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	waitExitCode(t, m)

	deadline := time.Now().Add(3 * time.Second)
	for {
		n := runtime.NumGoroutine()
		if n <= base {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("进程退出后仍有 goroutine 泄漏: %d > baseline %d（Stop 后 %v）",
				n, base, time.Since(baseline))
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// ---------- worker 路径注入 ----------

// TestManagerSetPathStopped SetPath 在 Stopped 状态生效、空串被忽略。
func TestManagerSetPathStopped(t *testing.T) {
	m := New("")
	if got := m.Path(); got != DefaultWorkerPath {
		t.Fatalf("空构造应回退默认路径, got %q", got)
	}

	m.SetPath("")
	if got := m.Path(); got != DefaultWorkerPath {
		t.Fatalf("空串应被忽略, got %q", got)
	}

	const want = `C:\fake\ninfer.exe`
	m.SetPath(want)
	if got := m.Path(); got != want {
		t.Fatalf("Stopped 状态 SetPath 应生效, want %q, got %q", want, got)
	}

	m2 := New("anything")
	m2.SetPath("")
	if got := m2.Path(); got != "anything" {
		t.Fatalf("空串不应覆盖已有路径, got %q", got)
	}
}

// TestManagerSetPathRunning SetPath 在 Running 状态不生效（不影响已启动进程）。
func TestManagerSetPathRunning(t *testing.T) {
	exe := buildFake(t, fakeSlowSrc)
	m := New(exe)
	m.Grace = 300 * time.Millisecond
	if err := m.Start(defaultCfg()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if m.State() != Running {
		t.Fatalf("Start 后状态 = %v, want Running", m.State())
	}

	m.SetPath(`C:\other\ninfer.exe`)
	if got := m.Path(); got != exe {
		t.Fatalf("Running 状态 SetPath 不应生效, want %q, got %q", exe, got)
	}

	if err := m.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	waitExitCode(t, m)
}
