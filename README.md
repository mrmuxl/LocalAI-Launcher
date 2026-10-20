# NInfer Launcher

NInfer 和llama.cpp 推理引擎的 Windows 桌面启动器（推理程序可执行文件名任意）：Fyne v2 中文图形界面，替代手工执行 `docs/scripts/windows/start-ninfer.ps1`，用表单完成参数配置、启动/停止进程与日志实时查看。

## 构建

需要 Go 1.26（依赖 `fyne.io/fyne/v2`）：

```sh
make build-launcher
```

产物为 `bin/LocalAI-launcher.exe`。

## 运行

1. 双击运行 `bin/LocalAI-launcher.exe`。窗口标题「LocalAI 启动器」，终端输出版本横幅 `LocalAI-launcher v0.1.0`。
2. 左栏双引擎 tab：**NInfer**（33 字段，对齐 start-ninfer.bat 默认配方）与 **llama.cpp**（46 字段，对齐 Qwen3.8-27B-GSQ_S-MTP.cmd 启动脚本），各自独立表单、独立推理程序行与配置。每个 tab 的「服务」组下有「推理程序」行指定该引擎的 worker 可执行文件（文件名任意）：输入路径后回车，或点「打开」弹文件选择窗口；llama.cpp tab 的「模型文件」「视觉模型」行同样有「打开」按钮。文件选择窗口为独立窗口（可用鼠标拖拽边角拉大），双击目录进入、双击文件选择，并记忆上次所在目录（持久化到 worker.json，下次打开直接定位）。启动时若已有存档则直接采用，否则静默使用默认路径并在操作反馈中提示，不会自动弹窗。所选路径持久化，下次启动直接使用。
3. 在当前 tab 的表单中填写配置项，点「保存配置」持久化、「重新加载」重新读取；点「启动」按当前 tab 的引擎拉起 worker，日志面板实时显示输出（保留最近 500 行）；「停止」先发 3 秒宽限等待自行退出，超时后强杀。

- 单实例：Windows 下使用命名互斥锁 `Global\NinferLauncher`，第二个实例弹中文提示并退出（其他平台该锁为 no-op）。

## 配置路径

全部配置文件便携式存放于启动器 exe 同级目录（整个目录拷走即整套配置，与推理程序位置无关）：

- NInfer：启动器 exe 同级目录下的 `ninfer.json`（33 字段）；
- llama.cpp：启动器 exe 同级目录下的 `llama.json`（46 字段）；
- worker 路径存档：启动器 exe 同级目录下的 `worker.json`。

- 文件不存在：加载使用内置默认值，不报错；
- JSON 损坏：加载报错，不覆盖原文件；
- 点「保存配置」时自动创建目录并写完整配置。

## worker 路径

启动器 exe 同级目录下的 `worker.json`：引导存档（记录两个引擎的推理程序在哪）。`workerPath` 存 NInfer 引擎的推理程序路径（文件名任意），`llamaWorkerPath` 存 llama.cpp 引擎的推理程序路径（文件名任意）；在对应 tab 的「推理程序」行输入路径回车或点「打开」选择时写入，仅保存可执行文件绝对路径，下次启动直接采用。存档指向的文件被删除时，启动会静默回退默认路径并提示，可重新指定。

## 与 server 的关系

- 启动器只是管理前端，不含任何服务端/推理逻辑，服务仍由 NInfer 推理程序提供。
- 点「启动」时按「argv 构造契约」（见 `internal/argv`）把表单配置逐字翻译成 worker 命令行参数（固定段 + 条件段 + 缓存段）再启动进程；GUI 字段与 argv flag 一一对应。
- 本仓库不修改 worker 源码；`docs/scripts/windows/*.ps1` 启动脚本与启动器是同一服务包的两条等价启动途径。
