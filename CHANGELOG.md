# 更新日志 (v1.5.3)

### 2026.10.15 - 原生 OIDC / Pocket ID 单点登录、任务批量设置多类型混选重构、表格分页数独立持久化与演示模式全面加固

🎉 **新增功能与体验升级**
* **原生 OIDC / Pocket ID 单点登录支持 (Feature & Auth & Security)**：
  - **主流 OIDC 协议原生集成**：支持基于 OpenID Connect (OIDC) 规范的单点登录 (SSO) 认证体系，深度适配并无缝兼容 Pocket ID、Keycloak、Authelia 等自建与主流开源 IAM 身份认证平台（#191）；
  - **双轨登录架构支持**：支持传统管理员账密与 OIDC 单点登录双轨并行，支持在系统设置中一键开启/停用 OIDC 并配置 Client ID、Client Secret 及 Endpoints；
  - **全链路安全防护与会话闭环**：前端提供一键快捷跳转授权，后端严格校验 PKCE / State 防 CSRF 攻击，全自动置换授权码并完成用户身份识别与 JWT 签名签发；
  - **单点登录与两步验证指南**：官方文档中心补充发布《两步验证 (OTP) 与单点登录 (OIDC) 完整安全指南》，涵盖 Pocket ID 详细配置、回调重定向 URI 与排障步骤。
* **任务多类型混选批量参数设置重构 (Tasks & Feature & UI)**：
  - **多类型混选智能兼容**：重构任务批量设置交互，全面解决白虎应用 (App)、仓库同步 (Repo) 与普通脚本 (Task) 在多选/全选批量操作时的参数冲突与类型歧义问题（#192）；
  - **模块化结构分类定义**：
    - **通用参数 (Common)**：超时时间、批量标签管理、调度重试与随机延迟、全量环境变量注入（支持全部任务类型）；
    - **运行时环境 (Language)**：运行时多语言环境 Mise（独立分类，按所选任务支持情况智能生效）；
    - **执行位置 (Agent)**：统一执行位置 Agent 节点（严格限制仅对普通脚本任务生效，并在界面提供明确的适用性标签与禁用说明）；
  - **批量操作栏与动画平滑优化**：移除多余的 CLI Hooks 指令设置，优化小屏与桌面端批量浮动操作栏的布局与交互。
* **表格每页条数独立持久化与分页体验升级 (Pagination & Web & UI)**：
  - **解耦全局 PageSize 依赖**：移除站点设置中对全局 `pageSize` 的强约束，各表格页面通过 `localStorage` 独立持久化记住用户的每页条数偏好；
  - **9 大核心表格全量覆盖**：定时任务、环境变量/机密、标签管理、执行历史以及系统所有日志 Tab 均实现独立的每页分页偏好记忆，互不干扰；
  - **卡片内嵌式底栏排版统一**：修复环境变量页面与任务列表页面的分页条脱落浮空缺陷，全量统一为紧凑内嵌式底栏设计，消除多余空隙与深浅色底色割裂；
  - **Inter 字体排版与精炼视觉**：页码及数据指示器全面应用 `Inter` 字体栈配合 `tabular-nums` 等宽特性，移除冗余的区间文本显示，呈现更加紧凑高级的现代微排版。
* **演示模式全方位安全加固 (Security & Demo)**：
  - **敏感配置与服务全链路锁定**：在演示模式（`BH_DEMO_MODE=true`）下全面拦截并禁用敏感接口与写操作：禁用 OTP 两步验证与 OIDC 单点登录配置、禁用 OpenAPI 接口调用与凭据重置、禁用 MCP 远程协议与客户端连接、禁用容器内存缓存优化与调度核心参数调整；
  - **前端交互状态防范**：前端表单输入项自动禁用并标明“演示模式禁用”醒目标记，彻底杜绝公开演示站点数据外泄与参数篡改风险。
* **视觉体验与流畅度综合打磨 (Perf & Style & Fix)**：
  - **暗色模式 Anti-FOUC 极速探测**：增加内联 Anti-FOUC 探测脚本，在 HTML 解析首帧毫秒级应用 `dark` 类，彻底根治暗色主题下页面刷新瞬间闪白的顽疾；
  - **弹窗动效去冗与极速响应**：移除 Dialog / AlertDialog 拖沓的缩放动画（`zoom-in-95`/`zoom-out-95`）与延迟，优化遮罩层渲染，使弹窗弹出与关闭极速干脆；
  - **应用市场多分支源容错**：仅在目标分支为 `main` 时优先读取 GitHub Pages 静态预编译源，避免 `dev` 等开发分支配置失效；
  - **UI 细节优化**：优化通知渠道卡片测试发送图标为纸飞机 `Send`，优化运行环境配置弹窗视觉与字体排版。

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

  直接运行附件中的安装程序 `BaihuPanel-Setup-v1.5.3-windows-amd64.exe` 完成安装，程序将自动完成桌面快捷方式创建、开机自启配置与 `127.0.0.1 baihu.local` 域名绑定。安装完成后通过桌面图标或任务栏右下角托盘图标即可直接打开 `http://baihu.local:38052`。

* **方式二：二进制单文件运行（免安装/便携版）**

  解压下载好的 `.zip` 压缩包（如 `baihu-windows-amd64.zip`），进入解压目录并在 PowerShell 中运行：

  ```powershell
  .\baihu.exe server
  ```

---

**访问面板：**
* 启动后访问：`http://baihu.local:38052` 或 `http://localhost:38052` (或单文件默认端口 `http://localhost:8052`)
* **默认账号**：用户名 `admin`，密码见面板首次启动时的控制台日志。
