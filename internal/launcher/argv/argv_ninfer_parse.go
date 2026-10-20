// Package argv 提供 launcher 的命令行参数解析能力。
// 仅依赖标准库，解析失败时返回携带具体信息的错误。
package argv

import (
	"errors"
	"fmt"
	"strconv"
)

// Cfg 保存从命令行解析出的全部启动参数。
type Cfg struct {
	// Model 指定要启动的推理模型名称（--model）。
	Model string
	// Device 指定推理设备（如 cpu / gpu），默认由模型配置决定（--device）。
	Device string
	// Host 指定服务监听地址（--host）。
	Host string
	// Port 指定服务监听端口（--port）。
	Port int
	// Threads 指定推理使用的线程数（--threads）。
	Threads int
}

// ErrHelp 表示用户通过 -h / --help 请求帮助信息。
var ErrHelp = errors.New("help requested")

// Parse 解析命令行参数 args（不含程序名）。
// 支持的 flag：--model=、--device=、--host=、--port=、--threads=、-h、--help。
// 未知 flag 返回包含该 flag 名称的错误；数值解析失败返回同时包含
// flag 名与原始字符串值的错误；-h / --help 返回 (Cfg{}, ErrHelp)。
func Parse(args []string) (Cfg, error) {
	var cfg Cfg

	for _, arg := range args {
		switch {
		case arg == "-h" || arg == "--help":
			return Cfg{}, ErrHelp
		case hasPrefix(arg, "--model="):
			cfg.Model = arg[len("--model="):]
		case hasPrefix(arg, "--device="):
			cfg.Device = arg[len("--device="):]
		case hasPrefix(arg, "--host="):
			cfg.Host = arg[len("--host="):]
		case hasPrefix(arg, "--port="):
			raw := arg[len("--port="):]
			n, err := strconv.Atoi(raw)
			if err != nil || n < 1 || n > 65535 {
				return Cfg{}, fmt.Errorf("invalid --port value %q: must be an integer in 1..65535", raw)
			}
			cfg.Port = n
		case hasPrefix(arg, "--threads="):
			raw := arg[len("--threads="):]
			n, err := strconv.Atoi(raw)
			if err != nil || n < 1 {
				return Cfg{}, fmt.Errorf("invalid --threads value %q: must be a positive integer", raw)
			}
			cfg.Threads = n
		default:
			return Cfg{}, fmt.Errorf("unknown flag %q", arg)
		}
	}

	return cfg, nil
}

// hasPrefix 判断 s 是否以 prefix 开头。
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
