// Package version 提供 LocalAI-launcher 的版本信息。
//
// 用途：启动时打印的版本横幅，供 cmd 入口调用；
// 构建产物与 Banner 统一使用 Version 常量，避免版本号散落多处。
package version

// Version 为构建产物与 Banner 统一使用的版本号。
const Version = "0.1.0"

// Banner 返回启动时打印的版本横幅。
func Banner() string {
	return "LocalAI-launcher v" + Version
}
