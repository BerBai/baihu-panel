# 更新日志 (v1.5.8)

### 2026.10.27 - 语言安装下载镜像加速与 Mise 运行时镜像源配置、已装应用更新检测与无损升级中枢、OpenAPI 文件管理能力开放与全站安全加固

🎉 **新增功能与体验升级**
* **语言安装下载镜像加速与 Mise 运行时镜像源配置 (Languages & Mise & Perf)**：
  - **新增语言下载镜像加速**：在新增语言弹窗中集成 GitHub 镜像选择功能，内置 `gh-proxy.com (推荐)`、`ghfast.top`、`ghproxy.net`、`github.moeyy.xyz` 及自定义代理前缀选项，彻底解决境内服务器通过 GitHub Releases 安装 Python/Node/Go 等运行时下载超时失败的痛点（[#200](https://github.com/engigu/baihu-panel/issues/200)）；
  - **全局默认加速源持久化与多系统命令注入**：支持将所选镜像持久化写入 Mise 全局配置（`MISE_URL_REPLACEMENTS`），后续自动化或终端安装全量生效；同时安装终端根据底层操作系统（Linux/Windows）自适应注入环境变量加速执行；
  - **加速选项卡双区分类管理**：将“加速”页面划分为“运行时下载加速 (Mise Runtime)”与“包管理器依赖源 (Package Registry)”两大区域，支持一键配置 Mise GitHub 加速、Node.js 运行时镜像（`MISE_NODE_MIRROR_URL`）、Go 官方国内镜像（`MISE_GO_DOWNLOAD_MIRROR`）及 Python 编译源码镜像（`PYTHON_BUILD_MIRROR_URL`）；
  - **系统环境变量无损提取优化**：修复优化 `mise_service` 中环境变量解析逻辑，支持带有空格与 JSON 结构的变量值完整无损还原。
* **已装应用更新检测与一键无损升级中枢 (Apps & Scheduler & Feature)**：
  - **系统级更新检测与手动检测**：支持已安装应用的版本更新自检，内置系统级每日定时巡检与手动立即检测，发现新版本时联动系统通知中心发送更新提醒（[#201](https://github.com/engigu/baihu-panel/pull/201)）；
  - **版本漂移检测与无损升级闭环**：支持实时校验当前运行版本与应用清单最新定义的版本差异，提供一键无损升级能力，自动原子重构调度实体、系统环境变量、关联任务与依赖环境，保障升级过程平滑无缝。
* **OpenAPI 文件管理接口开放与 404 兜底修复 (OpenAPI & Web & Fix)**：
  - **开放文件写入与保存接口**：OpenAPI 正式开放文件保存与写入端点，支持第三方工具及 CI/CD 自动化流水线远程下发管理文件（[#202](https://github.com/engigu/baihu-panel/pull/202)）；
  - **静态资源与 SPA 路由 404 兜底机制完善**：修复优化前端路由与静态资源兜底响应行为。
* **全站核心依赖安全加固与 Dependabot 漏洞清零 (Security & Dependencies)**：
  - **核心工具链安全升级**：升级 `vite` 至 `^7.3.7`（Web 端）与 `^6.4.5`（Docs 端），彻底消除 `launch-editor` 命令注入与路径校验绕过漏洞；
  - **全链路依赖漏洞修复**：同步升级与覆盖 `dompurify` (`^3.4.17`)、`zod` (`^4.6.7`)、`undici` (`^7.30.1`)、`fast-uri` (`^4.2.3`)、`serialize-javascript` (`^7.1.4`)，实现主系统与文档站点全量 `npm audit` 0 漏洞，GitHub Dependabot 安全告警全部闭环清零。

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

  直接运行附件中的安装程序 `BaihuPanel-Setup-v1.5.8-windows-amd64.exe` 完成安装，程序将自动完成桌面快捷方式创建、开机自启配置与 `127.0.0.1 baihu.local` 域名绑定。安装完成后通过桌面图标或任务栏右下角托盘图标即可直接打开 `http://baihu.local:38052`。

* **方式二：二进制单文件运行（免安装/便携版）**

  解压下载好的 `.zip` 压缩包（如 `baihu-windows-amd64.zip`），进入解压目录并在 PowerShell 中运行：

  ```powershell
  .\baihu.exe server
  ```

---

**访问面板：**
* 启动后访问：`http://baihu.local:38052` 或 `http://localhost:38052` (或单文件默认端口 `http://localhost:8052`)
* **默认账号**：用户名 `admin`，密码见面板首次启动时的控制台日志。
