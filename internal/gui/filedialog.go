// filedialog.go 自定义文件选择弹窗：独立窗口（可用鼠标拉大）、
// 双击操作（双击目录进入、双击文件选择）、记忆上次所在目录。
//
// 线程模型：全部方法在主线程调用（App 按钮回调）；onPick 回调由调用方
// 决定是否派发后台 goroutine。
package gui

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// doubleClickWindow 两次单击视为双击的最大间隔。
const doubleClickWindow = 500 * time.Millisecond

// dirEntry 是文件列表的一行（目录或文件）。
type dirEntry struct {
	name  string
	isDir bool
}

// filePicker 自定义文件选择窗口。
type filePicker struct {
	win     fyne.Window
	pathBox *widget.Entry // 当前目录（可编辑，回车跳转）
	list    *widget.List
	entries []dirEntry
	cur     string   // 当前目录
	exts    []string // 允许的扩展名（含点，小写）；空 = 全部文件

	lastID    int                           // 上次单击的列表项（双击检测）
	lastSelAt time.Time                     // 上次单击时间
	onPick    func(path string, dir string) // 双击文件：完整路径 + 所在目录
}

// newFilePicker 构建文件选择窗口（主线程调用，不自动显示）。
// start 为起始目录；exts 为扩展名过滤（如 []string{".gguf"}，空 = 不过滤）；
// onPick 在双击文件时回调（主线程），dir 为文件所在目录（供记忆）。
func newFilePicker(app fyne.App, title, start string, exts []string, onPick func(path, dir string)) *filePicker {
	p := &filePicker{
		exts:   exts,
		onPick: onPick,
		lastID: -1,
	}
	p.win = app.NewWindow(title)
	p.win.Resize(fyne.NewSize(720, 520)) // 独立窗口，可用鼠标拖拽边角拉大

	p.pathBox = widget.NewEntry()
	p.pathBox.OnSubmitted = func(text string) {
		if st, err := os.Stat(text); err == nil && st.IsDir() {
			p.enter(text)
		} else {
			p.pathBox.SetText(p.cur)
		}
	}

	upBtn := widget.NewButton("上级", func() { p.enter(filepath.Dir(p.cur)) })

	p.list = widget.NewList(
		func() int { return len(p.entries) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			e := p.entries[i]
			l := o.(*widget.Label)
			if e.isDir {
				l.SetText("[目录] " + e.name)
			} else {
				l.SetText(e.name)
			}
		},
	)
	p.list.OnSelected = func(i widget.ListItemID) {
		p.onListSelected(i)
	}

	p.win.SetContent(container.NewBorder(
		container.NewHBox(upBtn, p.pathBox),
		nil, nil, nil,
		p.list,
	))
	p.enter(start)
	return p
}

// show 显示窗口（主线程）。
func (p *filePicker) show() { p.win.Show() }

// onListSelected 处理列表单击：同项 500ms 内再次单击视为双击——
// 目录进入、文件回调 onPick 并关闭窗口。
func (p *filePicker) onListSelected(i widget.ListItemID) {
	now := time.Now()
	if int(i) == p.lastID && now.Sub(p.lastSelAt) < doubleClickWindow {
		p.lastID = -1
		p.activate(i)
		return
	}
	p.lastID = int(i)
	p.lastSelAt = now
}

// activate 执行双击动作（主线程）。
func (p *filePicker) activate(i widget.ListItemID) {
	if int(i) >= len(p.entries) {
		return
	}
	e := p.entries[i]
	if e.isDir {
		p.enter(filepath.Join(p.cur, e.name))
		return
	}
	path := filepath.Join(p.cur, e.name)
	dir := p.cur
	p.win.Close()
	p.onPick(path, dir)
}

// enter 进入指定目录并刷新列表（主线程）。
func (p *filePicker) enter(dir string) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Printf("解析目录 %s 失败: %v", dir, err)
		return
	}
	items, err := os.ReadDir(abs)
	if err != nil {
		log.Printf("读取目录 %s 失败: %v", abs, err)
		return
	}

	var dirs, files []dirEntry
	for _, it := range items {
		if it.IsDir() {
			dirs = append(dirs, dirEntry{name: it.Name(), isDir: true})
			continue
		}
		if p.matchExt(it.Name()) {
			files = append(files, dirEntry{name: it.Name()})
		}
	}
	sort.Slice(dirs, func(a, b int) bool { return strings.ToLower(dirs[a].name) < strings.ToLower(dirs[b].name) })
	sort.Slice(files, func(a, b int) bool { return strings.ToLower(files[a].name) < strings.ToLower(files[b].name) })
	p.entries = append(dirs, files...)
	p.cur = abs
	p.lastID = -1

	p.pathBox.SetText(abs)
	p.list.Refresh()
}

// matchExt 判断文件名是否匹配扩展名过滤（空 exts = 全部通过）。
func (p *filePicker) matchExt(name string) bool {
	if len(p.exts) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(name))
	for _, e := range p.exts {
		if ext == strings.ToLower(e) {
			return true
		}
	}
	return false
}
