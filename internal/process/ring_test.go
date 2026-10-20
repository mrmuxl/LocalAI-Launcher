package process

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// TestRingAppendBelowCapacity 未满时按追加顺序保留全部行。
func TestRingAppendBelowCapacity(t *testing.T) {
	r := NewRingLog(4)
	for i := 0; i < 3; i++ {
		r.Append(fmt.Sprintf("line-%d", i))
	}
	if r.Len() != 3 {
		t.Fatalf("Len = %d, want 3", r.Len())
	}
	got := r.Snapshot()
	want := []string{"line-0", "line-1", "line-2"}
	if len(got) != len(want) {
		t.Fatalf("快照长度 = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("快照[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestRingOverflowDropsOldest 计划契约：500 行容量，追加 501 行后
// 最旧行（line-001）被丢弃，保留最近 500 行且顺序不变。
func TestRingOverflowDropsOldest(t *testing.T) {
	r := NewRingLog(500)
	for i := 1; i <= 501; i++ {
		r.Append(fmt.Sprintf("line-%03d", i))
	}
	if r.Len() != 500 {
		t.Fatalf("满溢后 Len = %d, want 500", r.Len())
	}
	got := r.Snapshot()
	if len(got) != 500 {
		t.Fatalf("快照长度 = %d, want 500", len(got))
	}
	if got[0] != "line-002" {
		t.Fatalf("最旧行 = %q, want line-002（line-001 应被丢弃）", got[0])
	}
	if got[499] != "line-501" {
		t.Fatalf("最新行 = %q, want line-501", got[499])
	}
	for i := range got {
		if want := fmt.Sprintf("line-%03d", i+2); got[i] != want {
			t.Fatalf("快照[%d] = %q, want %q（序列不连续）", i, got[i], want)
		}
	}
}

// TestRingDefaultCapacity 非法容量回退到默认 500，满溢行为一致。
func TestRingDefaultCapacity(t *testing.T) {
	r := NewRingLog(0)
	if cap(r.buf) != DefaultRingSize {
		t.Fatalf("默认容量 = %d, want %d", cap(r.buf), DefaultRingSize)
	}
	for i := 1; i <= 501; i++ {
		r.Append(strconv.Itoa(i))
	}
	if r.Len() != DefaultRingSize {
		t.Fatalf("满溢后 Len = %d, want %d", r.Len(), DefaultRingSize)
	}
	got := r.Snapshot()
	if got[0] != "2" || got[499] != "501" {
		t.Fatalf("满溢两端 = %q..%q, want 2..501", got[0], got[499])
	}
}

// TestRingConcurrentAppendAndSnapshot 并发写 + 并发快照（-race 门禁），
// 满溢后快照必须恰好是 1000 个不同行中的最后 500 行：
// 若环形覆盖逻辑有 bug（写错槽位），500 个槽位装 500 行必现重复。
func TestRingConcurrentAppendAndSnapshot(t *testing.T) {
	r := NewRingLog(500)

	stop := make(chan struct{})
	var snapWG sync.WaitGroup
	snapWG.Add(1)
	go func() {
		defer snapWG.Done()
		for {
			select {
			case <-stop:
				return
			default:
				_ = r.Snapshot()
			}
		}
	}()

	var wg sync.WaitGroup
	const perG = 100
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				r.Append(fmt.Sprintf("g%d-%d", g, i))
			}
		}(g)
	}
	wg.Wait()
	close(stop)
	snapWG.Wait()

	if r.Len() != 500 {
		t.Fatalf("并发满溢后 Len = %d, want 500", r.Len())
	}
	got := r.Snapshot()
	if len(got) != 500 {
		t.Fatalf("快照长度 = %d, want 500", len(got))
	}
	seen := make(map[string]bool, 500)
	for _, line := range got {
		parts := strings.SplitN(line, "-", 2)
		if len(parts) != 2 || !isDigits(parts[0][1:]) || !isDigits(parts[1]) {
			t.Fatalf("快照出现非法行 %q", line)
		}
		if seen[line] {
			t.Fatalf("快照出现重复行 %q：环形覆盖逻辑错误", line)
		}
		seen[line] = true
	}
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
