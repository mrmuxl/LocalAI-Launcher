[简体中文](#简体中文) | [English](#english)

## 简体中文

> **启动前必读：运行包和模型必须使用英文路径，完整路径中不能含中文字符，包括用户名、上级目录和文件名。**

# KVMem v0.18.0 正式版 · NInfer · Windows RTX 30/40/50

Windows x64 正式发布，包含 NInfer 服务端、KVMem、内嵌 Web UI 和运行依赖。使用 NInfer v3 `.ninfer` 模型；模型另行准备，本包仅包含 NInfer 后端。

## 相比 RC3 的变化

- 0.18.0 正式版启动配方于 2026-10-13 更新：专用和统一 Windows 入口默认 sink=12288 tokens（12K）、recent=1024 tokens（1K，16页）。
- 64 是 NInfer 支持的最小值；0–63 会被拒绝，以保护需要继续追加的最新 KV 页。
- 引擎及运行依赖复用 RC3 的同一二进制。更新启动脚本、说明、版本元数据和 SHA256；旧 RC3 保留供追溯。此次参数调整未重新测量长上下文质量或性能。

## 相比 RC2 的主要变化

- **工具调用约束**：默认接入原生工具调用约束，覆盖 Chat Completions、Responses 和 Anthropic Messages 的工具选择与 strict schema，并与 thinking、JSON 输出及 MTP/ngram 推测解码配合；只发布完整工具调用。不支持的 schema 仍会明确报错。
- **RAM/SSD 分层 KV 管理**：活跃和闲置历史共用 RAM/SSD 配额，完整原生 K/V、量化附属数据及 MTP KV 可下沉到 SSD；恢复时按实际选中的内容读取。增加小记录合并写入和空闲区间索引。SSD 默认关闭，显式指定 `-DiskPath` 后启用。
- **LRFU 历史加权与 GQA max**：Windows 启动入口默认开启 LRFU，`lambda=0.5`、`top-frac=0.9`，GQA 聚合采用 `max`；sink 默认 12288 tokens（12K），recent 默认 1024 tokens（1K，16页）。每个请求只提交一次历史更新，续接、编辑、重复请求和同进程历史恢复使用对应的历史状态。
- **客户端兼容修复**：接受图片 `detail=low/high/original` 并映射为 `auto`；Responses 的 reasoning item 返回 `summary=[]`，保留原始推理文本，避免 DSH 将原始思考重复当作摘要处理。
- **默认参数调整**：专用 `start-ninfer.ps1` 的图片 token 上限由 RC2 的 256 改为 1024。统一 `start-backend.ps1` 默认 GPU budget 改为 32768；NInfer 快捷配方仍为 B/R=36864/16384。

RC2 的多卡按层分配、权重量化和 KV 格式支持继续保留。LRFU 来自参考实现的历史加权思路，本次没有加入 decode 期间的持续 attention score 累积，也没有自适应预算或配额混合。

## 选择下载包

| 显卡 | 下载包 | 大小 |
|---|---|---|
| RTX 30 系 / sm86 | [kvmem-v0.18.0-ninfer-windows-rtx30-sm86.zip](https://github.com/kvmem/kvmem/releases/download/v0.18.0/kvmem-v0.18.0-ninfer-windows-rtx30-sm86.zip) | 971.6 MiB |
| RTX 40 系 / sm89 | [kvmem-v0.18.0-ninfer-windows-rtx40-sm89.zip](https://github.com/kvmem/kvmem/releases/download/v0.18.0/kvmem-v0.18.0-ninfer-windows-rtx40-sm89.zip) | 964.5 MiB |
| RTX 50 系 / sm120a | [kvmem-v0.18.0-ninfer-windows-rtx50-sm120a.zip](https://github.com/kvmem/kvmem/releases/download/v0.18.0/kvmem-v0.18.0-ninfer-windows-rtx50-sm120a.zip) | 859.0 MiB |

需要支持 CUDA 13.2 的 NVIDIA 驱动及 Microsoft Visual C++ 2015–2022 x64 运行库。CUDA 运行依赖随包提供，无需安装 CUDA Toolkit 或 Visual Studio。NVFP4 模型权重仅适用于 RTX 50 系包。

## 快速开始

1. 根据显卡下载 ZIP，完整解压到英文路径。
2. 将匹配的 `.ninfer` 模型放到包根目录并命名为 `model.ninfer`，或通过 `-Model` 指定路径。
3. 在解压目录打开 PowerShell，查看显卡，再启动服务。示例选择编号 0；也可用完整 GPU UUID 代替编号。

```powershell
nvidia-smi -L

powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\windows\start-ninfer.ps1 `
    -Model 'D:\models\model.ninfer' -Gpu '0'
```

就绪后打开 `http://127.0.0.1:18200/`。OpenAI 兼容 API 地址为 `http://127.0.0.1:18200/v1`。保持终端开启，Ctrl+C 停止服务；参数可直接追加到命令中。

默认开启 CPU 视觉、自适应 MTP4 和 ngram31，模型须包含对应权重。缺少视觉权重时追加 `-NoVision`；缺少 MTP 权重时追加 `-MtpDrafts 0`，同时关闭默认自适应 MTP 和 ngram。

## 模型准备

优先下载模型作者提供的 **NInfer v3 `.ninfer` 版本**。完整模型包含所需配套资源，无需另外下载 JSON 或 mmproj；开启视觉或 MTP 时须有对应权重。

只有 GGUF 时，参考 [ninfer-all 的转换工具与使用指南](https://github.com/iamwavecut/ninfer-all/blob/8319e8f512477d0cfdef89e16c17c1ba9f9bc7b3/docs/weight-conversion.md) 转换。

## 默认启动参数

以下为包内 **`scripts/windows/start-ninfer.ps1` 专用入口**的默认值；1k＝1024 tokens。

| 参数 | 默认值 |
|---|---|
| 总 context | 200k / 204800 |
| GPU budget / Gen reserve | 36k / 36864；16k / 16384 |
| 默认单次输出上限 | 16k / 16384 |
| Main / MTP KV | 双 INT8 |
| MTP | 自适应开启，草稿上限 4 |
| Ngram drafting / min-match | 31 / 12 |
| 视觉 | 默认开启，CPU；`-NoVision` 关闭 |
| 图片 token 上限 | 1024 |
| Thinking budget | 8192 |
| Host KV payload 预算 | 12288 MiB / 12 GiB |
| Prefill chunk | 256 tokens |
| Sink / recent | 12288 / 1024 tokens（12K / 1K） |
| 保留会话 / 并发 | 1 / 1 |
| LRFU | 开启，lambda=0.5，top-frac=0.9 |
| GQA 聚合 | max |
| SSD 分层 / 设备自动调参 | 均关闭 |

这套配方在 RTX 5060 Ti 16GB 的 Qwen3.8-27B GSQ-RCO IQ3_S 图文/MTP 模型上完成加载和文本请求验证；不保证其他模型或显存容量都能装下。客户端可覆盖单次输出长度，实际输出也受剩余 context 限制。`-Context` 设置上下文上限，不保证任意内存条件下均能达到该长度。

KV 格式仍支持 `bf16`、`int8`、`fp8`、`nvfp4`、`k8v4`、`rk8v4`、`rk4v4`、`rk4v4-e8` 和 `rk2v4-e8`。支持文本与图片，视频暂不支持。

统一 `start-backend.ps1 -Backend ninfer` 使用另一套默认参数，B=32768；直接运行 worker 也不会自动采用上述配方。请使用 `start-ninfer.ps1` 获得表中默认值。

## LRFU 和检索参数

LRFU 用当前 query 与历史 KV 块的相关性更新历史分数，再对剩余预算中的块排序。sink、recent 和必要的媒体保留规则继续生效，计入现有预算。

`top-frac=0.9` 控制允许累加本轮相关性分数的命中范围，**不是保留 90% 的 KV，也不是预算比例**。GQA max 对同一 KV head 对应的 query heads 取最大相关性。

```powershell
# 覆盖历史加权参数
& .\scripts\windows\start-ninfer.ps1 -Gpu '0' `
    -LrfuLambda 0.5 -LrfuTopFraction 0.9 -ScoreAggregation max

# 关闭历史加权，并使用 mean 聚合
& .\scripts\windows\start-ninfer.ps1 -Gpu '0' -NoLrfu -ScoreAggregation mean
```

缩小 B 时，也要确保 sink/recent 能放进预算；可用 `-SinkTokens` 和 `-RecentTokens` 覆盖。用 `-DryRun` 查看完整启动参数。

## 可选 SSD 分层

```powershell
& .\scripts\windows\start-ninfer.ps1 -Gpu '0' `
    -DiskPath 'D:\kvmem-cache' -DiskMiB 4096
```

只指定 `-DiskPath` 时，磁盘预算默认 4096 MiB。`-HostMiB` 与 `-DiskMiB` 分别约束原生 KV 等受管理数据的 RAM/SSD 配额，**不是进程总内存上限**；状态、统计、暂存和操作系统文件缓存仍占用额外内存。RAM 必须容纳活动工作集与换页余量，增大磁盘预算不能代替这部分 RAM。

SSD 文件仅供当前服务进程使用，**不支持重启恢复或跨后端 KV 迁移**。LRFU 历史元数据随同进程会话恢复保留。磁盘满或 I/O 失败会按请求边界报错，不提前丢弃唯一有效副本。详见 [RAM/SSD 分层说明](https://github.com/kvmem/kvmem/blob/938c2f1968b0283e7d7dd143918a9f3ff866dd4d/docs/active-kv-tiering.md)。

## 多卡启动

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\windows\start-ninfer.ps1 `
    -Gpu '0,1'
```

支持最多 8 张相同 CUDA 计算能力的显卡，省略层数时按可用显存自动分配。手动分配时在当前 PowerShell 会话直接调用脚本，例如 64 层模型：

```powershell
Set-ExecutionPolicy -Scope Process Bypass
& .\scripts\windows\start-ninfer.ps1 -Gpu '0,1' -StageLayers 40,24
```

每张卡填写一个正整数，顺序对应 `-Gpu`，合计等于模型层数。使用匹配架构的包；MTP 位于列表中的第一张卡，视觉默认在 CPU。也可用逗号连接 GPU UUID。双卡能力沿用 RC2，本次发布包未重新做多卡性能测试。

## RC3 已完成的性能实测（2026-10-13）

以下直接沿用 RC3 在 **Windows、RTX 5060 Ti 16GB、Qwen3.8-27B GSQ-RCO IQ3_S 图文/MTP 模型**上的已完成测试。正式版保留同一引擎二进制。

| 项目 | RC3 实测结果 | 范围 |
|---|---:|---|
| 日常文本回答生成 | 50.08 token/s | 正式 3 轮聚合 |
| 日常图片回答生成 | 55.90 token/s | 正式 3 轮聚合 |
| 日常 HTML/SVG 代码生成 | 64.54 token/s | 正式 3 轮聚合，每轮 512 tokens |
| 日常完整任务耗时 | 30.10 秒 | 3 轮总耗时的中位数 |
| 长历史工具轮次生成 | 53.24 token/s | 32 轮聚合，含最终代码轮 |
| 长历史最终代码生成 | 44.87 token/s | 末轮服务端原生速度 |
| 长历史最终首字等待 | 16.87 秒 | 到首个思考或正文片段 |
| 长历史请求总耗时 | 855.10 秒（约 14.25 分钟） | 33 次请求耗时相加，不含加载 |
| 最终输入 / 输出 | 261,522 / 512 tokens | 接近 256K；1K=1024 |
| 请求完成 / 工具轮次前缀复用 | 33/33 完成；32/32 命中 | 单次完整长历史测试 |

测试配方：context=262144，B/R=36864/16384，Main/MTP KV 均为 INT8，GDN state=FP16；固定 MTP3（关闭自适应），ngram=0；prefill=512，thinking=128，每次最多 512 输出 tokens；GPU resident 视觉（图片上限 1024）；LRFU(.5/.9)、GQA max；Host=12288 MiB，SSD 关闭，会话/并发=1/1，device profile=off；temperature=1、top_p=.95、top_k=20、min_p=0、seed=42。接口为流式 Chat Completions。

日常任务先预热 1 轮，再测量 3 轮，包含约 12K 文本、896×896 图片和 HTML/SVG 生成；代码轮原生速度为 62.01、68.31、63.25 token/s。长历史采用注入的模拟 `read_file` 结果，33 次请求均完成，末轮只为适配模板开销删去 54 行重复填充语句。聚合速度按总输出 tokens / 总服务端生成耗时计算，包含思考；工具执行时间未计入，也未据此验证代码完整性或通用召回率。

整卡显存采样峰值约 14.12 GiB。长历史阶段系统可用物理 RAM 最低降至 **8.78 MiB**，存在明显内存压力，可能影响首字等待和总耗时；Host 预算不是进程总内存上限。

完整报告、原始计时汇总、配置与服务日志：[RC3 性能记录](https://github.com/kvmem/kvmem/releases/download/v0.18.0/RC3-performance-results-20261013.zip)。

## 验证范围与已知限制

- 三包均完成 Release 编译、CUDA 架构、运行依赖、干净 PATH 下启动帮助入口和 PowerShell 5.1 启动参数检查；ZIP 重新解压并校验包内文件。
- RTX 50 系预编译包已实际部署到 **RTX 5060 Ti**，完成模型加载、健康检查、Responses 请求及 DSH Responses 适配器的短流式请求验证。RTX 30/40 系尚无对应硬件的实际推理验证。
- 本次没有重新执行性能测试、完整图像推理矩阵或 SSD 压力测试。分层实现的功能与历史测试范围见 [验收记录](https://github.com/kvmem/kvmem/blob/938c2f1968b0283e7d7dd143918a9f3ff866dd4d/docs/active-kv-tiering-validation-20261010.md)；物理内存紧张或冷 SSD 下的性能仍需验证。
- [RC2 的 256k 任务二结果](https://github.com/kvmem/kvmem/releases/tag/v0.18.0-ninfer-rc2) 属于旧版本及其专用压力配置，不代表 0.18.0 默认配方的性能。
- 中文路径问题沿用 RC2 的路径要求。`--kvmem-verify-transfers` 是默认关闭的诊断选项，历史 256k 压测曾出现内存不足，本次未重新启用。

## 校验和与反馈

ZIP 校验文件：[SHA256SUMS-ninfer-only.txt](https://github.com/kvmem/kvmem/releases/download/v0.18.0/SHA256SUMS-ninfer-only.txt)。各包内 `SHA256SUMS` 用于校验解压后的文件。

反馈请附显卡、驱动、包版本、模型、修改后的参数和报错；日志中请去除私有对话。

本次默认启动配方修订来自提交 [`fd776139e62742fbcf7d03b08730c33db8f3cea0`](https://github.com/kvmem/kvmem/commit/fd776139e62742fbcf7d03b08730c33db8f3cea0)；`v0.18.0` 标签保留首次发布源码快照。更新包中的 `launcher_source_commit` 记录脚本修订，原生引擎和 RC3 性能证据保持不变。

发布源码来自 `release/0.18.0`，提交 [`ea476bac448e17fb50a8b694cd65cde949eaca9b`](https://github.com/kvmem/kvmem/commit/ea476bac448e17fb50a8b694cd65cde949eaca9b)；引擎二进制仍来自 RC3 的 `938c2f1968b0283e7d7dd143918a9f3ff866dd4d`。NInfer 基础提交为 `8319e8f512477d0cfdef89e16c17c1ba9f9bc7b3`，叠加本仓库集成补丁；完整来源见包内 `BUILD-INFO.json`。本发布包仅含 NInfer 后端。

## 致谢

感谢原始引擎与社区分支的维护者和贡献者：

- [Neroued/ninfer](https://github.com/Neroued/ninfer)：原始引擎。
- [Don-Chad/ninfer-3090](https://github.com/Don-Chad/ninfer-3090)、[ashalliants/ninfer-3090](https://github.com/ashalliants/ninfer-3090)及[WarlaxZ](https://github.com/WarlaxZ)：3090分支与后续引擎工作。
- [UDPSendToFailed/ninfer-4090](https://github.com/UDPSendToFailed/ninfer-4090)、[TertiumOrganum1/ninfer-3090](https://github.com/TertiumOrganum1/ninfer-3090)：适配与优化。
- [iamwavecut/ninfer-all](https://github.com/iamwavecut/ninfer-all)：多分支整合，RC2与RC3采用的NInfer源码基础。
- [Ryan-gsq的Windows与GSQ-RCO分支](https://github.com/Ryan-gsq/ninfer-16g-5070ti-5080-5090-qwen3.8-27b-gsq-rco)：RC1源码基础，以及Windows与GSQ-RCO适配工作。
- [MirkoCovizzi/ninfer-rtx5090-mobile](https://github.com/MirkoCovizzi/ninfer-rtx5090-mobile)：自适应MTP。
- [Wallawalla47/ninfer-custom](https://github.com/Wallawalla47/ninfer-custom)的Ian Ranson及remesis等贡献者：快速prefill和ngram等工作。

也感谢[IMGillusion/ninfer-disk-kv](https://github.com/IMGillusion/ninfer-disk-kv)、[tmark00/ninfer](https://github.com/tmark00/ninfer)、[gzenz/ninfer](https://github.com/gzenz/ninfer)及其他PR作者。详细署名见[RC1上游贡献记录](https://github.com/Ryan-gsq/ninfer-16g-5070ti-5080-5090-qwen3.8-27b-gsq-rco/blob/6e9189c1089f6968e29fe2e1c51b4596e2346eed/docs/maintainer/consolidated-line.md)，许可证随包保留。

LRFU历史加权与GQA max实现参考 [melis2023/kvmem-actkv](https://github.com/melis2023/kvmem-actkv)。第三方许可证随包保留；cuBLAS DLL按包内CUDA EULA分发，仅供本包应用使用，不适用NInfer的Apache-2.0许可证。


## English

[简体中文](#简体中文) | English

> **Before starting: Use English paths for the runtime package and model. The full path must not contain Chinese characters, including the username, parent directories, and filename.**

# KVMem v0.18.0 Release · NInfer · Windows RTX 30/40/50

Stable Windows x64 release with the NInfer server, KVMem, an embedded Web UI, and runtime dependencies. Models must be supplied separately in NInfer v3 `.ninfer` format. These packages contain only the NInfer backend.

## Changes from RC3

- The 2026-10-13 update to stable 0.18.0 sets dedicated/unified Windows launcher defaults to sink=12288 tokens (12K) and recent=1024 tokens (1K, 16 pages).
- 64 is the NInfer minimum. Values 0–63 are rejected to protect the newest native KV page for further appends.
- Retains the exact RC3 engine binaries and dependencies; updates launchers, documentation, release metadata and SHA256. The RC3 release remains available for traceability. Long-context quality and performance were not remeasured for this setting.

## Main changes from RC2

- **Tool-call constraints**: enabled by default for tool selection and strict schemas across Chat Completions, Responses, and Anthropic Messages. Constraints compose with thinking, JSON output, and MTP/ngram decoding; only complete calls are published. Unsupported schemas still produce explicit errors.
- **RAM/SSD KV tiering**: active and idle histories share RAM/SSD quotas. Native K/V, quantization metadata, and MTP KV can move to SSD; restoration reads the selected content. Adds batched small-record writes and an indexed free-space allocator. SSD is disabled unless `-DiskPath` is specified.
- **LRFU history and GQA max**: Windows launchers enable LRFU with `lambda=0.5`, `top-frac=0.9`, and `max` GQA aggregation. Sink defaults to 12288 tokens (12K), recent to 1024 tokens (1K, 16 pages). Each request commits history once; continuations, edits, repeated requests, and in-process restoration use the corresponding history state.
- **Client compatibility**: image `detail=low/high/original` is accepted and mapped to `auto`. Responses reasoning items use `summary=[]` while preserving raw reasoning text, preventing DSH from treating that text as a duplicate summary.
- **Defaults**: the dedicated launcher's image token limit increases from RC2's 256 to 1024. The unified launcher's GPU budget is now 32768; the dedicated NInfer recipe remains B/R=36864/16384.

RC2's multi-GPU layer allocation, weight quantization, and KV formats remain available. This LRFU implementation does not accumulate attention scores continuously during decode, add adaptive budgets, or mix retrieval quotas.

## Choose a package

| GPU | Package | Size |
|---|---|---|
| RTX 30 / sm86 | [kvmem-v0.18.0-ninfer-windows-rtx30-sm86.zip](https://github.com/kvmem/kvmem/releases/download/v0.18.0/kvmem-v0.18.0-ninfer-windows-rtx30-sm86.zip) | 971.6 MiB |
| RTX 40 / sm89 | [kvmem-v0.18.0-ninfer-windows-rtx40-sm89.zip](https://github.com/kvmem/kvmem/releases/download/v0.18.0/kvmem-v0.18.0-ninfer-windows-rtx40-sm89.zip) | 964.5 MiB |
| RTX 50 / sm120a | [kvmem-v0.18.0-ninfer-windows-rtx50-sm120a.zip](https://github.com/kvmem/kvmem/releases/download/v0.18.0/kvmem-v0.18.0-ninfer-windows-rtx50-sm120a.zip) | 859.0 MiB |

Requires a CUDA 13.2-compatible NVIDIA driver and the Microsoft Visual C++ 2015–2022 x64 Redistributable. CUDA dependencies are bundled; neither CUDA Toolkit nor Visual Studio is required. NVFP4 model weights require the RTX 50 package.

## Quick start and model preparation

Extract the entire ZIP for your GPU to an English path. Obtain a compatible NInfer v3 model, place it in the package root as `model.ninfer`, or specify `-Model`. No separate JSON or mmproj is needed for a complete model. To convert GGUF, see the [ninfer-all conversion guide](https://github.com/iamwavecut/ninfer-all/blob/8319e8f512477d0cfdef89e16c17c1ba9f9bc7b3/docs/weight-conversion.md).

Open PowerShell in the extracted directory:

```powershell
nvidia-smi -L

powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\windows\start-ninfer.ps1 `
    -Model 'D:\models\model.ninfer' -Gpu '0'
```

Replace `0` with your GPU index or UUID. Open `http://127.0.0.1:18200/` after loading. The OpenAI-compatible API is `http://127.0.0.1:18200/v1`. Keep the terminal open; Ctrl+C stops the server.

CPU vision, adaptive MTP4, and ngram31 are enabled by default and require the corresponding model weights. Add `-NoVision` if vision weights are missing. Add `-MtpDrafts 0` if MTP weights are missing; this also disables the implicit adaptive MTP and ngram settings.

## Default startup parameters

Defaults below apply to the dedicated **`scripts/windows/start-ninfer.ps1`** entry point. Here, 1k means 1024 tokens.

| Parameter | Default |
|---|---|
| Context | 200k / 204800 |
| GPU budget / Gen reserve | 36k / 36864; 16k / 16384 |
| Maximum output per request | 16k / 16384 |
| Main / MTP KV | INT8 for both |
| MTP | Adaptive, up to 4 draft tokens |
| Ngram drafts / min-match | 31 / 12 |
| Vision / image token limit | CPU, enabled / 1024 |
| Thinking budget | 8192 |
| Host KV payload budget | 12288 MiB / 12 GiB |
| Prefill chunk | 256 tokens |
| Sink / recent | 12288 / 1024 tokens (12K / 1K) |
| Retained sessions / concurrency | 1 / 1 |
| LRFU | Enabled, lambda=0.5, top-frac=0.9 |
| GQA aggregation | max |
| SSD tiering / automatic device tuning | Disabled |

This recipe loaded a Qwen3.8-27B GSQ-RCO IQ3_S vision/MTP model and completed text requests on an RTX 5060 Ti 16GB. Other models or VRAM capacities may require adjustments. Clients can override the default output limit; remaining context also limits output. A context ceiling does not guarantee that every memory configuration can reach it.

Available KV formats remain `bf16`, `int8`, `fp8`, `nvfp4`, `k8v4`, `rk8v4`, `rk4v4`, `rk4v4-e8`, and `rk2v4-e8`. Text and images are supported; video is not.

The unified `start-backend.ps1 -Backend ninfer` entry point has different defaults, including B=32768. Running the worker directly also does not apply this recipe.

## Retrieval and optional SSD settings

LRFU updates historical block scores using current query relevance, then ranks blocks within the remaining budget. Existing sink/recent and required-media retention rules remain effective. `top-frac=0.9` controls the score-update hit range; it does **not** retain 90% of KV or allocate 90% of the budget. GQA max takes the maximum relevance among query heads sharing a KV head.

```powershell
# Override retrieval settings
& .\scripts\windows\start-ninfer.ps1 -Gpu '0' `
    -LrfuLambda 0.5 -LrfuTopFraction 0.9 -ScoreAggregation max

# Disable history weighting and use mean aggregation
& .\scripts\windows\start-ninfer.ps1 -Gpu '0' -NoLrfu -ScoreAggregation mean

# Enable SSD with a 4096 MiB disk budget
& .\scripts\windows\start-ninfer.ps1 -Gpu '0' `
    -DiskPath 'D:\kvmem-cache' -DiskMiB 4096
```

If B is reduced, adjust `-SinkTokens` and `-RecentTokens` so the retention windows fit. Use `-DryRun` to inspect the arguments. Specifying only `-DiskPath` uses a 4096 MiB disk budget.

Host/disk quotas cover managed payloads rather than total process memory. State, statistics, staging buffers, and OS file caches use additional memory. RAM must still accommodate active working sets and swapping headroom. SSD files are **process-local: no restart recovery or cross-backend KV migration**. LRFU metadata remains attached to in-process history. Disk-full and I/O failures fail requests at safe boundaries without discarding the sole valid copy. See the [tiering guide](https://github.com/kvmem/kvmem/blob/938c2f1968b0283e7d7dd143918a9f3ff866dd4d/docs/active-kv-tiering.md).

## Multiple GPUs

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\windows\start-ninfer.ps1 `
    -Gpu '0,1'

# Manual allocation for a 64-layer model, in the current PowerShell session
Set-ExecutionPolicy -Scope Process Bypass
& .\scripts\windows\start-ninfer.ps1 -Gpu '0,1' -StageLayers 40,24
```

Supports up to 8 GPUs with the same CUDA compute capability. Layers are allocated automatically by available VRAM unless specified. Manual counts must be positive, match GPU order, and sum to the model's layer count. Use the package matching the GPU architecture. MTP uses the first GPU; vision defaults to CPU. Comma-separated UUIDs are also supported. Multi-GPU support is retained from RC2; this release did not repeat multi-GPU performance tests.

## Completed RC3 performance measurements (2026-10-13)

These reuse completed RC3 measurements on **Windows, RTX 5060 Ti 16GB, and the Qwen3.8-27B GSQ-RCO IQ3_S vision/MTP model**. The stable release retains the exact engine binary.

| Metric | RC3 measured result | Scope |
|---|---:|---|
| Daily text response generation | 50.08 token/s | 3 measured runs, aggregate |
| Daily image response generation | 55.90 token/s | 3 measured runs, aggregate |
| Daily HTML/SVG code generation | 64.54 token/s | 3 measured runs, 512 output tokens each |
| Complete daily task | 30.10 s | Median full-task duration across 3 runs |
| Long-history tool rounds | 53.24 token/s | 32 rounds aggregated, including final code |
| Final long-history code generation | 44.87 token/s | Native server rate for the final request |
| Final time to first token | 16.87 s | First reasoning or text fragment |
| Long-history request time | 855.10 s (about 14.25 min) | Sum of 33 requests, excluding model load |
| Final input / output | 261,522 / 512 tokens | Near 256K; 1K=1024 |
| Completion / prefix reuse | 33/33 completed; 32/32 tool-round hits | One complete long-history run |

Measured settings: context=262144, B/R=36864/16384; INT8 Main/MTP KV, FP16 GDN state; fixed MTP3, adaptive off, ngram=0; prefill=512, thinking=128, output cap=512; GPU-resident vision with image cap=1024; LRFU(.5/.9), GQA max; Host=12288 MiB, SSD off, sessions/concurrency=1/1, device profile off; temperature=1, top_p=.95, top_k=20, min_p=0, seed=42. Requests use streaming Chat Completions.

The daily workload used one warmup plus three measured runs with about 12K text tokens, an 896×896 image and HTML/SVG generation. Native code-generation samples were 62.01, 68.31 and 63.25 token/s. The long workload injected simulated `read_file` results; all 33 requests completed. The final fixture removed 54 repeated filler lines to accommodate template overhead. Aggregate rates use total output tokens / total server generation time and include reasoning. They exclude real tool execution and do not establish code completeness or general retrieval quality.

Sampled whole-card VRAM peaked at about 14.12 GiB. Available system RAM fell to **8.78 MiB** during the long workload; that memory pressure may affect first-token and end-to-end times. The Host payload budget is not a process memory limit.

Full report, timing summaries, configuration and server logs: [RC3 performance records](https://github.com/kvmem/kvmem/releases/download/v0.18.0/RC3-performance-results-20261013.zip).

## Validation and limitations

- All packages passed Release builds, CUDA architecture and runtime dependency checks, clean-PATH help startup, PowerShell 5.1 launcher checks, and extraction/file checksum verification.
- The RTX 50 package was deployed on an **RTX 5060 Ti** and passed model loading, health, Responses requests, and a short streaming request through the installed DSH Responses adapter. RTX 30/40 hardware inference validation remains pending.
- No new performance benchmark, complete image-inference matrix, or SSD stress test was run for these packages. See the [tiering validation record](https://github.com/kvmem/kvmem/blob/938c2f1968b0283e7d7dd143918a9f3ff866dd4d/docs/active-kv-tiering-validation-20261010.md) for implementation-level evidence. Performance under physical RAM pressure or cold SSD conditions remains unverified.
- [RC2's 256k Task Two results](https://github.com/kvmem/kvmem/releases/tag/v0.18.0-ninfer-rc2) describe an earlier version and its stress-test settings, not the 0.18.0 default recipe.
- The RC2 English-path requirement still applies. The optional `--kvmem-verify-transfers` diagnostic remains disabled; a historical 256k test ran out of memory with it enabled, and it was not re-enabled here.

## Checksums, source, and feedback

Verify ZIPs with [SHA256SUMS-ninfer-only.txt](https://github.com/kvmem/kvmem/releases/download/v0.18.0/SHA256SUMS-ninfer-only.txt). Each archive contains `SHA256SUMS` for extracted files.

For reports, include the GPU, driver, package version, model, modified parameters, and error. Remove private conversations from logs.

This launcher-default revision comes from commit [`fd776139e62742fbcf7d03b08730c33db8f3cea0`](https://github.com/kvmem/kvmem/commit/fd776139e62742fbcf7d03b08730c33db8f3cea0). The v0.18.0 tag preserves the initial source snapshot; launcher_source_commit records the script revision. Native engine binaries and RC3 performance evidence are unchanged.

Release source: `release/0.18.0`, commit [`ea476bac448e17fb50a8b694cd65cde949eaca9b`](https://github.com/kvmem/kvmem/commit/ea476bac448e17fb50a8b694cd65cde949eaca9b). Engine binaries retain RC3 commit `938c2f1968b0283e7d7dd143918a9f3ff866dd4d`. NInfer base: `8319e8f512477d0cfdef89e16c17c1ba9f9bc7b3`, with this repository's integration patch. See `BUILD-INFO.json` inside the package for provenance. Only the NInfer backend is bundled.


## Acknowledgments

Thanks to the maintainers and contributors of the original engine and community branches:

- [Neroued/ninfer](https://github.com/Neroued/ninfer): the original engine.
- [Don-Chad/ninfer-3090](https://github.com/Don-Chad/ninfer-3090), [ashalliants/ninfer-3090](https://github.com/ashalliants/ninfer-3090), and [WarlaxZ](https://github.com/WarlaxZ): the 3090 branches and subsequent engine work.
- [UDPSendToFailed/ninfer-4090](https://github.com/UDPSendToFailed/ninfer-4090) and [TertiumOrganum1/ninfer-3090](https://github.com/TertiumOrganum1/ninfer-3090): adaptation and optimization.
- [iamwavecut/ninfer-all](https://github.com/iamwavecut/ninfer-all): integration of multiple branches and the NInfer source base used by RC2 and RC3.
- [Ryan-gsq's Windows and GSQ-RCO branch](https://github.com/Ryan-gsq/ninfer-16g-5070ti-5080-5090-qwen3.8-27b-gsq-rco): the RC1 source base, with Windows and GSQ-RCO adaptations.
- [MirkoCovizzi/ninfer-rtx5090-mobile](https://github.com/MirkoCovizzi/ninfer-rtx5090-mobile): adaptive MTP.
- [Wallawalla47/ninfer-custom](https://github.com/Wallawalla47/ninfer-custom), with Ian Ranson, remesis, and other contributors: fast prefill and ngram work.

Thanks also to [IMGillusion/ninfer-disk-kv](https://github.com/IMGillusion/ninfer-disk-kv), [tmark00/ninfer](https://github.com/tmark00/ninfer), [gzenz/ninfer](https://github.com/gzenz/ninfer), and other PR authors. See the [RC1 upstream contributor record](https://github.com/Ryan-gsq/ninfer-16g-5070ti-5080-5090-qwen3.8-27b-gsq-rco/blob/6e9189c1089f6968e29fe2e1c51b4596e2346eed/docs/maintainer/consolidated-line.md) for detailed attribution. Licenses are retained in the packages.

LRFU history weighting and GQA max follow the reference implementation in [melis2023/kvmem-actkv](https://github.com/melis2023/kvmem-actkv). Third-party licenses are retained. Bundled cuBLAS DLLs are distributed under the included CUDA EULA for use with this application, not under NInfer's Apache-2.0 license.

