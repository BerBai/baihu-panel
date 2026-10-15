# 更新日志 (v1.5.4)

### 2026.10.15 - MCP 双传输共存架构 (Streamable HTTP + SSE)、通用 OAuth 2.0 单点登录与全站 Inter 字体排版

🎉 **新增功能与体验升级**
* **MCP 协议双传输共存架构 (Streamable HTTP + HTTP/SSE) (Major Feature & AI & MCP)**：
  - **Streamable HTTP 单端点现行规范集成**：全面支持 2025-03-26 MCP 现行官方标准规范，单端点（`/open2api/v1/mcp`）统一通信，支持即时 JSON-RPC 响应与渐进式流式传输，会话基于标准请求头 `Mcp-Session-Id`，天然穿透企业网关与反向代理，无缝适配 OpenCode 2、Cursor、Claude Desktop、DSH、Cline 等现代 AI 客户端（#193）；
  - **HTTP + SSE 双端点兼容模式**：完整保留 2024-11-05 早期双端点规范（`GET /open2api/v1/mcp/sse` 与 `POST /open2api/v1/mcp/messages`），为仅支持旧版协议的客户端提供无缝向下兼容；
  - **底层共享与长连接防爆熔断**：两种传输通道共享同一个底层 `NewBaihuMCPServer` 业务与工具实例，统一复用现有 OpenAPI Bearer Token 及 `?token=` 鉴权；保留 SSE 活跃连接数限制（最多支持 3 个长连接会话并发维持），防止海量空闲长连接耗尽系统协程资源；
  - **前端双协议切换与配置指引**：系统设置中“AI 客户端连接 (MCP)”卡片新增 Streamable HTTP 与 HTTP + SSE 协议快捷切换器，支持动态显示对应端点 URL 并一键复制客户端配置 JSON；
  - **文档与测试全量就绪**：补齐完整的 MCP 自动化集成测试套件（15 个测试 100% PASS），同步更新官方接入指南文档。
* **通用 OAuth 2.0 单点登录协议集成与互联折叠面板 (Feature & Auth & UI)**：
  - **通用 OAuth 2.0 原生集成**：在原有 OIDC 基础上扩展支持通用 OAuth 2.0 协议标准（Authorization Code Flow），支持自定义授权端点、令牌端点、用户信息端点、作用域（Scopes）与用户唯一标识键（兼容 GitHub、Gitee、GitLab 等任意主流 OAuth 2.0 平台）；
  - **互联风格折叠配置卡片**：安全设置中单点登录配置升级为优雅的折叠抽屉卡片，并自动自适应协议类型展示对应配置项；
  - **登录页无感双模适配**：登录界面根据后端配置自动适配 OIDC 或 OAuth 2.0 快捷登录入口，展示定制平台名称与直观的授权跳转。
* **全局 Inter 现代字体与执行历史排版优化 (Style & UI)**：
  - **全站现代 Inter 字体栈统一**：全站默认英文与数字字体升级为现代 `Inter` 字体，强化数字 `tabular-nums` 等宽渲染，呈现精致克制的高级感排版；
  - **执行历史底栏排版统一**：移除执行历史列表多余的空白撑开，分页器与筛选状态直接吸附在表格底栏，全量与其他核心业务页面保持一致。

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

  直接运行附件中的安装程序 `BaihuPanel-Setup-v1.5.4-windows-amd64.exe` 完成安装，程序将自动完成桌面快捷方式创建、开机自启配置与 `127.0.0.1 baihu.local` 域名绑定。安装完成后通过桌面图标或任务栏右下角托盘图标即可直接打开 `http://baihu.local:38052`。

* **方式二：二进制单文件运行（免安装/便携版）**

  解压下载好的 `.zip` 压缩包（如 `baihu-windows-amd64.zip`），进入解压目录并在 PowerShell 中运行：

  ```powershell
  .\baihu.exe server
  ```

---

**访问面板：**
* 启动后访问：`http://baihu.local:38052` 或 `http://localhost:38052` (或单文件默认端口 `http://localhost:8052`)
* **默认账号**：用户名 `admin`，密码见面板首次启动时的控制台日志。
