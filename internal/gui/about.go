// about.go 构建左栏「说明」tab 的纯展示内容（主线程构造期调用）。
package gui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// newAboutContent 构建「说明」tab 内容：
// 「说明」大标题 → 项目地址（三条超链接）→ 推理引擎后端（段落）→ 使用要点（列表），
// 各分区以分隔线隔开、列表/段落相对小节标题缩进，整体套 Scroll 防小窗溢出。
func newAboutContent() fyne.CanvasObject {
	// 大标题：粗体 + 标题字号（theme.SizeNameHeadingText）
	title := widget.NewLabelWithStyle("说明", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.SizeName = theme.SizeNameHeadingText

	// 「项目地址」小节：三条项目超链接（URL 与文案沿用原值，本项目链接与右栏状态区 repoLink 一致），逐行加「• 」项目符号
	llamaURL, _ := url.Parse("https://github.com/ggml-org/llama.cpp")
	llamaLink := widget.NewHyperlink("llama.cpp 项目地址：https://github.com/ggml-org/llama.cpp", llamaURL)
	llamaLink.Wrapping = fyne.TextWrapWord

	kvmemURL, _ := url.Parse("https://github.com/kvmem/kvmem")
	kvmemLink := widget.NewHyperlink("NInfer (kvmem) 项目地址：https://github.com/kvmem/kvmem", kvmemURL)
	kvmemLink.Wrapping = fyne.TextWrapWord

	// 本项目地址（与右栏状态区 repoLink、README 一致）
	launcherURL, _ := url.Parse("https://github.com/mrmuxl/LocalAI-Launcher")
	launcherLink := widget.NewHyperlink("本项目 (LocalAI-Launcher) 项目地址", launcherURL)
	launcherLink.Wrapping = fyne.TextWrapWord

	projects := container.NewPadded(container.NewVBox(
		bulletRow(llamaLink),
		bulletRow(kvmemLink),
		bulletRow(launcherLink),
	))

	// 「推理引擎后端」小节：一段说明文字（换行 + 缩进）
	backend := container.NewPadded(wrapLabel(
		"kvmem 可以使用 llama.cpp 和 ninfer 作为推理引擎后端，官方项目预编译只有 ninfer 后端（经预编译包实测）。本启动器不支持 kvmem 版 llama.cpp 后端，仅支持官方 llama.cpp。"))

	// 「使用要点」小节：四条列表项（换行 + 缩进）
	points := container.NewPadded(container.NewVBox(
		wrapLabel("• 需要去相关目录下载必要的推理引擎程序，已经在上面列出来了"),
		wrapLabel("• 配置文件便携式存放于启动器 exe 同级目录（ninfer.json / llama.json / worker.json）"),
		wrapLabel("• 「启动」按当前 tab 的引擎拉起 worker，日志面板实时显示输出（保留最近 500 行）"),
		wrapLabel("• 「停止」等待 3 秒宽限后强杀"),
		wrapLabel("• Windows 下以 CREATE_NO_WINDOW 启动，不弹出推理引擎的 cmd 窗口"),
		wrapLabel("• Linux下还未测试，我认为Linux用systemd来管理更加合适"),
		wrapLabel("• 相关模型可以去https://www.modelscope.cn和https://huggingface.co下载"),
	))

	box := container.NewVBox(
		title,
		widget.NewSeparator(),
		sectionHeading("项目地址"),
		projects,
		widget.NewSeparator(),
		sectionHeading("推理引擎后端"),
		backend,
		widget.NewSeparator(),
		sectionHeading("使用要点"),
		points,
	)
	return container.NewScroll(box)
}

// sectionHeading 构建粗体小节标题（与表单分组标题同款样式）。
func sectionHeading(text string) *widget.Label {
	return widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
}

// bulletRow 构建超链接列表行：「• 」项目符号标签与超链接同行（HBox 间距由 theme.Padding 提供）。
func bulletRow(link *widget.Hyperlink) fyne.CanvasObject {
	return container.NewHBox(widget.NewLabel("•"), link)
}

// wrapLabel 构建可换行（fyne.TextWrapWord）的列表项/段落标签。
func wrapLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}
