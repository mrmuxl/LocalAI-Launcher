# LocalAI-Launcher

Windows 桌面推理引擎启动器。基于 Fyne v2 的中文图形界面，用于管理 NInfer 与 llama.cpp 两个推理引擎的配置、启动/停止与日志查看。推理引擎可执行文件名任意。

## 功能特性

- 双引擎 tab 图形界面，表单化参数配置与逐字段中文校验
- 一键启动/停止推理引擎，实时日志面板（双引擎独立环形缓冲，各保留最近 500 行）
- llama.cpp 支持单模型与 route 预设（models-preset）两种启动模式，支持 MoE 相关参数
- 便携式配置：全部配置文件存于 exe 同级目录，整目录拷贝即完成迁移
- 单实例保护（Windows 命名互斥锁，重复启动弹窗提示）
- 引擎后台静默启动（CREATE_NO_WINDOW，不弹出引擎 cmd 窗口）

## 环境要求

- 运行：Windows 10 及以上
- 构建：Go 1.26+（依赖 `fyne.io/fyne/v2`）

## 快速开始

### 构建

```sh
make build-launcher
```

产物为 `bin/LocalAI-launcher.exe`。发布构建用 `make build-release`（`-H windowsgui` 链接标志，双击 exe 不弹控制台/cmd 窗口，直接出界面）。版本经 ldflags 注入：取 `git describe`（有 tag 用 tag，无 tag 用短 commit hash，工作区脏加 `-dirty`），也可用 `make build-release VERSION=x.y.z` 显式指定。

### 运行

双击 `bin/LocalAI-launcher.exe` 即可。界面分为左右两栏：左栏为双引擎 tab 表单，右栏自上而下为状态区、操作行与日志区，中间分割条可拖动。右栏状态区显示运行状态（状态/PID/运行时长/最新日志）、版本号及 GitHub 项目链接与维护者邮箱。

![maingui](https://github.com/mrmuxl/LocalAI-Launcher/blob/main/docs/maingui.png)

## 使用说明

### 推理引擎路径

每个 tab 的「服务」组下「推理引擎」行指定该引擎 worker 可执行文件（文件名任意）。输入路径回车或点「打开」选择文件；llama.cpp tab 的「模型文件」「视觉模型」「聊天模板文件」行同样有「打开」按钮。路径持久化到 worker.json，下次启动直接采用；存档指向的文件被删除时静默回退默认路径并在界面提示。文件选择起始目录自动定位到上次所在目录。

### 启动模式（llama.cpp）

llama.cpp tab「服务」组下「启动模式」下拉框提供两种模式：

- 单模型（默认）：`-m` 直启
- route 预设（models-preset）：`--models-preset` 加载 model.ini，模型/视觉由预设文件决定

切到 route 模式时「模型文件」「视觉模型」行隐藏，显示「预设文件」行。

### 启动与停止

- 「启动」：按当前 tab 引擎拉起 worker
- 「停止」：等待 3 秒宽限后强杀（Windows 标准库不支持向子进程发送优雅退出信号，宽限期为固定等待）

## 配置文件

全部配置文件便携式存放于启动器 exe 同级目录：

- `ninfer.json`：NInfer 配置
- `llama.json`：llama.cpp 配置（含启动模式）
- `worker.json`：两引擎 worker 路径存档

行为规则：

- 文件不存在：加载使用内置默认值，不报错
- JSON 损坏：加载报错，不覆盖原文件
- 保存时自动创建目录并写完整配置

## 架构说明

- 启动器是纯管理前端，不含任何服务端/推理逻辑，服务由推理引擎提供
- 启动时把表单配置逐字翻译为引擎命令行参数（GUI 字段与参数一一对应）
- 本仓库不修改引擎源码

## 联系方式

- GitHub: https://github.com/mrmuxl/LocalAI-Launcher
- 邮箱: xl@zhiwo.net

## License

本项目基于 Apache License 2.0 发布，详见 LICENSE 文件。
