# 更新日志 (v1.5.2)

### 2026.10.12 - 容器PageCache自适应回收与点杀、tini僵尸进程回收、树形目录批量加载与小屏代码编辑优化

🎉 **新增功能与体验升级**
* **容器与进程生命周期健壮性增强 (Docker & Core & Fix)**：
  - **tini 僵尸进程回收**：Docker 镜像入口升级引入 `tini` 作为 1 号 init 进程，彻底解决长时间多任务频繁派生可能产生的孤儿进程堆积与僵尸进程未回收问题（#188）；
  - **移除冗余 data/bin 软链目录**：移除无意义的 `data/bin` 软链创建，改由终端环境直连白虎主程序所在物理目录；Dockerfile 固化 `BH_ROOT_DIR=/app`，彻底消除无用目录与文件系统残留。
* **任务编辑与工作目录批量展开 (Tasks & API & Fix)**：
  - **批量子目录预获取**：新增批量获取下级文件树接口，解决定时任务编辑弹窗中复杂深层工作目录初次展开失败及多层目录逐级请求卡顿问题（#189）。
* **任务执行日志与排版渲染优化 (Logs & Web & Fix)**：
  - **Emoji 字符重叠修复**：修复脚本执行日志终端输出包含特定 Emoji 符号时与紧随其后的文字相互重叠排版的缺陷（#190）；
  - **日志操作栏一键复制**：日志卡片详情顶部工具栏新增便捷复制全文按钮，统一并发池与终端监控头部视觉风格。
* **移动端与代码编辑器体验重构 (Editor & Mobile & UI)**：
  - **移动端小屏横滑与下拉防击穿**：代码编辑器注入 `touchAction` 与 `overscroll-contain`，彻底消除移动端小屏左右拖动时误触发全局下拉刷新击穿的痛点；
  - **自动换行/横滑自由切换**：工具栏新增轻量切换按钮（“换行”/“横滑”），满足手机与桌面不同场景的代码浏览与编辑需求；
  - **滚动条小屏常驻**：小屏设备下 Monaco 编辑器横向与纵向滚动条保持常驻可见并调优高对比度滑块。
* **容器 Page Cache 智能自适应回收与点杀优化 (MemOpt & Cgroup & Docker)**：
  - **深层环境目录遍历与核心二进制点杀**：将容器缓存目录扫描最大深度从 3 层扩展至 6 层，彻底解决 Mise 环境下多语言核心可执行程序（如 `node`、`python`、`dotnet` 位于 `installs/<tool>/<ver>/bin/<tool>` 第 4 层）被深度限制遗漏导致巨额 Active Page Cache 常驻问题；在遍历中智能跳过 `include`、`share`、`doc`、`man` 等无关头文件与文档目录，大幅减少磁盘 I/O 开销与遍历耗时；
  - **联动 cgroup v2 原生快速回收机制**：在系统启动就绪阶段与常规工作区回收前优先调用 `TryReclaimCgroupMemory`（写入 `/sys/fs/cgroup/memory.reclaim`），实现毫秒级批量卸载容器可回收 Slab 目录项与非映射 Page Cache；
  - **高水位与半载智能防护策略**：建立容器内存高警戒水位（默认 80%）与半载活跃缓存重型任务（60%）双轨自适应回收策略，充裕状态下静默放行保证读写性能，紧张状态自动介入；
  - **任务结束防抖与尾部合并**：挂载 EventBus 任务结束事件监听器，引入 `DirDebouncer` 尾部合并防抖器，同目录多任务高频/连续结束时自动合并为最后一次统一回收，避免高频 I/O 抖动；
  - **自适应回收日志合并对比**：重构并精简自适应内存回收日志，单行透明输出前后全量指标对比（总占用百分比、文件缓存总额、活跃/非活跃变化、Docker 官方口径真实物理占用）。
* **监控大盘视觉与现代排印重构 (Monitor & Style & UI)**：
  - **字体层级与现代排印重构**：监控面板全面统一引入 Inter 字体，核心指标与图表强化数字加粗渲染与 `tabular-nums` 等宽特性，消除数值跳变抖动；针对移动端小屏设备强化粗体渲染；
  - **色彩降噪与仪表盘卡片重构**：收敛监控面板整体色彩，去除仪表盘与 Worker 卡片中的重复与冗余指标，强化视觉质感与信息层级清晰度；
  - **图表右上角指标徽章与视觉细节**：图表卡片右上角增加动态指标徽章，优化坐标轴与折线/柱状图视觉展现。
* **文档与更新大纲体验升级 (Docs & Changelog)**：
  - 文档中心更新日志的时间与版本号提升为二级大纲，优化 VitePress 侧边目录导航的层级切换与快速跳转体验。

---

> 💡 **提示**：出于安全及环境隔离考虑，推荐使用 Docker/Compose 部署方式。[镜像地址](https://github.com/engigu/baihu-panel/pkgs/container/baihu)

### 🐳 方式一：Docker 部署 (推荐)
[部署文档](https://github.com/engigu/baihu-panel?tab=readme-ov-file#%E5%BF%AB%E9%80%9F%E9%83%A8%E7%BD%B2)

---

### 🚀 方式二：单文件/安装包部署 (Linux / Windows)
从当前 Release 的附件中下载对应架构和平台的部署压缩包（Linux 为 `.tar.gz`，Windows 为安装包 `.exe` 或 `.zip`）。

#### 🐧 Linux 平台

**1. 安装前置依赖 `mise`**

单文件直接运行依赖宿主机系统环境，请务必先安装 [mise](https://mise.jdx.dev/getting-started.html) 供任务调度及环境管理使用：

```bash
curl https://mise.run | sh
export PATH="~/.local/share/mise/bin:~/.local/share/mise/shims:$PATH"
```

**2. 运行面板**

```bash
tar -xzvf baihu-linux-amd64.tar.gz
chmod +x baihu-linux-amd64
./baihu-linux-amd64 server
```

#### 🪟 Windows 平台

**1. 安装前置依赖**

* **安装 `mise`**（用于统一依赖和运行时环境管理）：

  在 PowerShell 中运行以下命令使用 `winget` 安装：
  ```powershell
  winget install jdx.mise
  ```

* **安装 `pwsh`**（PowerShell 7.6+，用于执行后台任务）：

  白虎面板在 Windows 下运行任务和工具链强依赖 PowerShell 7+。请参考 [微软官方 PowerShell 安装文档](https://learn.microsoft.com/zh-cn/powershell/scripting/install/install-powershell-on-windows?view=powershell-7.6) 安装，或通过 `winget` 快捷安装：
  ```powershell
  winget install Microsoft.PowerShell
  ```

**2. 运行面板**

* **方式一：使用 GUI 安装包安装（推荐）**

  直接运行附件中的安装程序 `BaihuPanel-Setup-v1.5.2-windows-amd64.exe` 完成安装，程序将自动完成桌面快捷方式创建、开机自启配置与 `127.0.0.1 baihu.local` 域名绑定。安装完成后通过桌面图标或任务栏右下角托盘图标即可直接打开 `http://baihu.local:38052`。

* **方式二：二进制单文件运行（免安装/便携版）**

  解压下载好的 `.zip` 压缩包（如 `baihu-windows-amd64.zip`），进入解压目录并在 PowerShell 中运行：

  ```powershell
  .\baihu.exe server
  ```

---

**访问面板：**
* 启动后访问：`http://baihu.local:38052` 或 `http://localhost:38052` (或单文件默认端口 `http://localhost:8052`)
* **默认账号**：用户名 `admin`，密码见面板首次启动时的控制台日志。
