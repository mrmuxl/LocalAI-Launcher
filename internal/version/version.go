// Package version 提供 LocalAI-launcher 的版本信息。
//
// 用途：启动时打印的版本横幅，供 cmd 入口调用；
// 构建产物与 Banner 统一使用 Version 常量，避免版本号散落多处。
package version

// Version 为构建产物与 Banner 统一使用的版本号。
// 默认值仅作兜底（go test / go run）；正式构建经 Makefile ldflags -X 注入
// git describe 版本（有 tag 用 tag，无 tag 用短 commit hash）。
var Version = "0.1.0"

// Banner 返回启动时打印的版本横幅。
func Banner() string {
	return "LocalAI-launcher v" + Version
}
