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
	"runtime"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
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
	nameBox *widget.Entry // 底部文件名框（单击文件回填，可手输后点「选择」）
	list    *widget.List
	entries []dirEntry
	cur     string   // 当前目录
	exts    []string // 允许的扩展名（含点，小写）；空 = 全部文件

	places      []placeItem // 左侧位置列表（磁盘 + 常用目录）
	placesList  *widget.List
	placeSelIdx int // 左侧当前选中项（-1 = 无）

	lastID    int                           // 上次单击的列表项（双击检测）
	lastSelAt time.Time                     // 上次单击时间
	onPick    func(path string, dir string) // 确认选择：完整路径 + 所在目录
}

// placeItem 左侧位置列表项：显示名 + 跳转目标目录。
type placeItem struct {
	name string
	path string
}

// newFilePicker 构建文件选择窗口（主线程调用，不自动显示）。
// 布局对齐 fyne 内置文件弹窗：顶部（上级 + 目录路径框）/ 中间文件列表 /
// 底部（文件名框 + 取消/选择按钮）。start 为起始目录；exts 为扩展名过滤
// （如 []string{".gguf"}，空 = 不过滤）；onPick 在确认选择时回调（主线程），
// dir 为文件所在目录（供记忆）。
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

	p.nameBox = widget.NewEntry()
	p.nameBox.SetPlaceHolder("选中文件后显示名称，也可直接输入")

	p.list = widget.NewList(
		func() int { return len(p.entries) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			e := p.entries[i]
			l := o.(*widget.Label)
			if e.isDir {
				l.SetText(e.name + "/")
			} else {
				l.SetText(e.name)
			}
		},
	)
	p.list.OnSelected = func(i widget.ListItemID) {
		p.onListSelected(i)
	}

	p.places = buildPlaces()
	p.placeSelIdx = -1
	p.placesList = widget.NewList(
		func() int { return len(p.places) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(p.places[i].name)
		},
	)
	p.placesList.OnSelected = func(i widget.ListItemID) {
		if int(i) >= len(p.places) {
			return
		}
		p.placeSelIdx = int(i)
		if p.places[i].path != p.cur {
			p.enter(p.places[i].path)
		}
	}

	cancelBtn := widget.NewButton("取消", func() { p.win.Close() })
	confirmBtn := widget.NewButton("选择", func() { p.confirm() })
	confirmBtn.Importance = widget.HighImportance

	nameRow := container.NewHBox(widget.NewLabel("文件名："), p.nameBox)
	btnRow := container.NewHBox(layout.NewSpacer(), cancelBtn, confirmBtn)
	bottom := container.NewVBox(nameRow, btnRow)

	p.win.SetContent(container.NewBorder(
		container.NewHBox(upBtn, p.pathBox),
		bottom,
		newFixedWidth(p.placesList, 150), nil,
		p.list,
	))
	p.enter(start)
	return p
}

// buildPlaces 枚举左侧位置列表：Windows 枚举全部盘符，再加主目录/桌面/
// 文档/下载（仅保留实际存在的项）。
func buildPlaces() []placeItem {
	var places []placeItem
	if runtime.GOOS == "windows" {
		for c := byte('A'); c <= 'Z'; c++ {
			drive := string(c) + `:\`
			if st, err := os.Stat(drive); err == nil && st.IsDir() {
				places = append(places, placeItem{name: drive, path: drive})
			}
		}
	} else if st, err := os.Stat("/"); err == nil && st.IsDir() {
		places = append(places, placeItem{name: "/", path: "/"})
	}
	if home, err := os.UserHomeDir(); err == nil {
		for _, it := range []struct{ label, sub string }{
			{"主目录", ""},
			{"桌面", "Desktop"},
			{"文档", "Documents"},
			{"下载", "Downloads"},
		} {
			path := home
			if it.sub != "" {
				path = filepath.Join(home, it.sub)
			}
			if st, err := os.Stat(path); err == nil && st.IsDir() {
				places = append(places, placeItem{name: it.label, path: path})
			}
		}
	}
	return places
}

// confirm 用文件名框当前内容确认选择（主线程）：文件存在且匹配过滤才回调。
func (p *filePicker) confirm() {
	name := strings.TrimSpace(p.nameBox.Text)
	if name == "" {
		return
	}
	path := filepath.Join(p.cur, name)
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return
	}
	dir := p.cur
	p.win.Close()
	p.onPick(path, dir)
}

// show 显示窗口（主线程）。
func (p *filePicker) show() { p.win.Show() }

// onListSelected 处理列表单击：单击文件回填底部文件名框；
// 同项 500ms 内再次单击视为双击——目录进入、文件确认选择。
func (p *filePicker) onListSelected(i widget.ListItemID) {
	now := time.Now()
	if int(i) == p.lastID && now.Sub(p.lastSelAt) < doubleClickWindow {
		p.lastID = -1
		p.activate(i)
		return
	}
	p.lastID = int(i)
	p.lastSelAt = now
	if int(i) < len(p.entries) && !p.entries[i].isDir {
		p.nameBox.SetText(p.entries[i].name)
	}
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
	p.nameBox.SetText(e.name)
	p.confirm()
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
	p.nameBox.SetText("")
	p.syncPlaceSelection(abs)
	p.list.Refresh()
}

// syncPlaceSelection 让左侧位置列表跟随当前目录：目录命中某个位置则选中，
// 否则取消选中（手动浏览子目录时左侧不误亮）。
func (p *filePicker) syncPlaceSelection(dir string) {
	for i, pl := range p.places {
		if pl.path == dir {
			if p.placeSelIdx != int(i) {
				p.placesList.Select(i) // 触发 OnSelected：目录相同不再 enter
			}
			return
		}
	}
	if p.placeSelIdx != -1 {
		p.placeSelIdx = -1
		p.placesList.UnselectAll()
	}
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
