// Package process 管理 ninfer worker 子进程生命周期与日志环形缓冲。
//
// 本包仅依赖标准库：不 import fyne，不接触任何 UI；
// 后台 goroutine 派生与收尾全部在进程退出后自然完成，不泄漏。
package process

import "sync"

// DefaultRingSize 计划「日志面板」约定的环形缓冲容量：最近 500 行。
// NewRingLog 传入 <= 0 的容量时回退到该默认值。
const DefaultRingSize = 500

// RingLog 是固定容量的行环形缓冲：满溢时丢弃最旧行，只保留最近 capacity 行。
// 全部方法线程安全，可在任意 goroutine 并发调用。
type RingLog struct {
	mu   sync.Mutex
	buf  []string // 容量 = len(buf)，槽位 0..cap-1
	head int      // 下一个写入槽位
	size int      // 当前有效行数（满时 = cap(buf)）
}

// NewRingLog 创建容量为 capacity 的环形缓冲；capacity <= 0 时取 DefaultRingSize(500)。
func NewRingLog(capacity int) *RingLog {
	if capacity <= 0 {
		capacity = DefaultRingSize
	}
	return &RingLog{buf: make([]string, capacity)}
}

// Append 追加一行；缓冲满时覆盖最旧行。
func (r *RingLog) Append(line string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	capacity := len(r.buf)
	r.buf[r.head] = line
	r.head = (r.head + 1) % capacity
	if r.size < capacity {
		r.size++
	}
}

// Snapshot 返回当前全部行的快照（最旧 → 最新），不修改缓冲。
func (r *RingLog) Snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	capacity := len(r.buf)
	out := make([]string, 0, r.size)
	start := (r.head - r.size + capacity) % capacity
	for i := 0; i < r.size; i++ {
		out = append(out, r.buf[(start+i)%capacity])
	}
	return out
}

// Len 返回当前有效行数。
func (r *RingLog) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.size
}
