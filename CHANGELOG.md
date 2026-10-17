# 更新日志 (v1.5.6)

### 2026.10.17 - 运行时PageCache控存策略重构、容器入口标准化与公共库抽取、OIDC领域层解耦及前端交互优化

🎉 **新增功能与体验升级**
* **运行时内存与 PageCache 控存策略重大优化 (Major Perf & MemOpt)**：
  - **动静分离保护运行时底座**：彻底排除 `/usr/local/bin/mise`、`/opt/mise-base` 等核心母盘环境目录，杜绝高频任务执行完毕后触发底层母盘 PageCache 频繁冷清理导致的磁盘冷读高发与 Block I/O 读流量暴涨痛点；基础运行底座仅保留在冷启动阶段执行单次初始化清理；
  - **任务结束局部精准清理与防抖冷却**：`OnTaskFinished` 回调严格仅清理该任务自身 `workDir` 产生的瞬时生成物，移除级联全盘扫描；并引入 30 秒目录级清理防抖冷却（`TaskFinishedCooldown`），彻底消减每 30 秒或更高频的短周期定时任务造成的磁盘 I/O 震荡；
  - **无限额容器安全阈值**：针对未配置内存上限（`LimitBytes == 0`）的 Docker 容器环境，将自动回收触发水位线动态提升至 512MB 以上，杜绝无意义的内存回收循环。
* **Docker 容器初始化入口公共库抽取与日志标准化 (Docker & DevOps & Style)**：
  - **公共入口函数库抽取**：新增 `docker-entrypoint-common.sh`，抽象并统一日志色彩、Banner 横幅、目录初始化、Mise 环境同步、PATH 与运行时环境注入、组件版本自检与 Bash 自动补全逻辑，大幅精简标准版（`docker-entrypoint.sh`）、极简版（`docker-entrypoint.minimal.sh`）与开发版（`docker-entrypoint-dev.sh`）的冗余维护成本；
  - **规范化版本横幅与属性对齐**：容器启动自动展示系统平台架构（如 `linux/amd64`）、程序版本号与构建时间戳，属性采用 6 个汉字等宽冒号竖直对齐；
  - **日志紧凑精简与彻底去重**：严格对齐 `build-go.sh`（`[GoBuild]`）规范格式，去除重复运行模式标题与多余空行，保持流水线般的极致紧凑与专业度。
* **OIDC / OAuth 2.0 领域层解耦重构 (Architecture & Auth)**：
  - **独立 OIDC 领域包抽离**：抽离独立的 `internal/oidc` 领域包（`config.go`、`provider.go`、`service.go`、`user.go`），实现配置解析、JWKS 发现缓存、端点构造、用户信息规范化与控制器彻底解耦；
  - **统一常量与配置读取入口**：统一常量与配置读取入口，大幅提升单点登录认证模块的内聚性与单元测试覆盖率。
* **前端任务交互细节与目录树选择优化 (Web & UI & Fix)**：
  - **仓库任务行执行时间修复**：修复仓库任务行（`RepoTaskRow.vue`）中最近执行时间的显示异常（#195）；
  - **目录树选择器初次加载优化**：任务工作目录选择树（`DirTreeSelect.vue`）初次展开时仅精准定位到当前目录，不再全量深度展开全部子目录，提升大目录下的交互流畅度（#195）；
  - **系统设置文案同步**：更新系统设置中容器内存优化与工作区清理的说明文案。
* **公共配置接口并发防抖与双端内存微缓存 (Perf & Network)**：
  - **前端 Promise In-Flight 去重与单页生命周期缓存**：针对系统设置页及其嵌入的多个子卡片（密码、2FA、单点登录）与登录页同一毫秒并发调用 `/api/v1/public/settings` 的网络浪费问题，在前端 API 层引入 Promise In-Flight 请求合并与 SPA 内存缓存，同一瞬间的多次调用直接复用首个请求，请求量直降 66%~75%，路由内二次读取 0 网络请求，并在保存设置时主动失效保证实时性；
  - **服务端常驻内存微缓存与 0ms 主动失效 (Public Cache)**：在 `internal/cache` 抽象独立的公共设置内存缓存（`public_cache.go`），在面板保存配置时通过 `SetSection` / `Set` 触发 0ms 瞬间主动失效；同时配置 30 秒安全兜底 TTL，兼顾了高并发下 0ms 亚毫秒级极速响应（<0.05ms）与运维手动直接改库场景下的自动感知与刷新。

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

  直接运行附件中的安装程序 `BaihuPanel-Setup-v1.5.6-windows-amd64.exe` 完成安装，程序将自动完成桌面快捷方式创建、开机自启配置与 `127.0.0.1 baihu.local` 域名绑定。安装完成后通过桌面图标或任务栏右下角托盘图标即可直接打开 `http://baihu.local:38052`。

* **方式二：二进制单文件运行（免安装/便携版）**

  解压下载好的 `.zip` 压缩包（如 `baihu-windows-amd64.zip`），进入解压目录并在 PowerShell 中运行：

  ```powershell
  .\baihu.exe server
  ```

---

**访问面板：**
* 启动后访问：`http://baihu.local:38052` 或 `http://localhost:38052` (或单文件默认端口 `http://localhost:8052`)
* **默认账号**：用户名 `admin`，密码见面板首次启动时的控制台日志。
