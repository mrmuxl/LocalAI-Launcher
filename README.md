# NInfer Launcher

NInfer 服务端（worker，`ninfer.exe`）的 Windows 桌面启动器：Fyne v2 中文图形界面，替代手工执行 `docs/scripts/windows/start-ninfer.ps1`，用表单完成参数配置、启动/停止进程与日志实时查看。

## 构建

需要 Go 1.26（依赖 `fyne.io/fyne/v2`）：

```sh
make build-launcher
```

产物为 `bin/ninfer-launcher.exe`。

## 运行

1. 双击运行 `bin/ninfer-launcher.exe`。窗口标题「NInfer 启动器」，终端输出版本横幅 `ninfer-launcher v0.1.0`。
2. 首次启动弹出文件选择框，指定 worker 可执行文件 `ninfer.exe`（仅显示 .exe）；所选路径持久化，下次启动直接使用，不再弹出。工具栏「选择 ninfer.exe…」可随时重新选择。
3. 在表单中填写 34 个配置项，点「保存配置」持久化、「重新加载」重新读取；点「启动」拉起 worker，日志面板实时显示输出（保留最近 500 行）；「停止」先发 3 秒宽限等待自行退出，超时后强杀。

- 单实例：Windows 下使用命名互斥锁 `Global\NinferLauncher`，第二个实例弹中文提示并退出（其他平台该锁为 no-op）。

## 配置路径

`%AppData%\NinferLauncher\config.json`（即 `os.UserConfigDir()/NinferLauncher/config.json`）。

- 文件不存在：加载使用内置默认值，不报错；
- JSON 损坏：加载报错，不覆盖原文件；
- 点「保存配置」时自动创建目录并写完整配置。

## worker 路径

`%AppData%\NinferLauncher\worker.json`：首次启动经文件选择框指定 `ninfer.exe` 时写入，仅保存可执行文件绝对路径；下次启动直接采用，工具栏「选择 ninfer.exe…」可随时重新选择。存档指向的文件被删除时，下一次启动会再次弹出选择框。

## 与 server 的关系

- 启动器只是管理前端，不含任何服务端/推理逻辑，服务仍由 `ninfer.exe` 提供。
- 点「启动」时按「argv 构造契约」（见 `internal/launcher/argv`）把表单配置逐字翻译成 worker 命令行参数（固定段 + 条件段 + 缓存段）再启动进程；GUI 字段与 argv flag 一一对应。
- 本仓库不修改 worker 源码；`docs/scripts/windows/*.ps1` 启动脚本与启动器是同一服务包的两条等价启动途径。
