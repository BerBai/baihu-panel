# 更新日志 (v1.5.5)

### 2026.10.16 - 多身份源单点登录并行共存架构、SSO 独立路由与凭据解耦、移动端响应式与视觉体验打磨

🎉 **新增功能与体验升级**
* **多身份源单点登录并行共存与认证架构深度重构 (Major Feature & Auth & SSO)**：
  - **多身份源并行启用 (Multi-IdP Coexistence)**：彻底打破标准 OIDC 与通用 OAuth 2.0 之间的互斥限制，支持企业级 OIDC（Pocket ID / Keycloak / Authentik）与主流通用 OAuth 2.0（GitHub / Gitee / GitLab 等）双通道同时独立开启与并行运行；
  - **配置与安全凭据物理解耦**：拆分为独立的启用开关（`oidc_enabled` 与 `oauth2_enabled`），各自的 Client ID、Client Secret（均受掩码脱敏与安全加密存储保护）、作用域（Scopes）、自动建号策略（`auto_create`）及默认角色物理隔离，支持在面板底部一键原子落库；
  - **独立路由组与接口重构**：新增通用 OAuth 2.0 独立路由组（`GET /api/v1/auth/oauth2/login`、`GET /api/v1/auth/oauth2/callback`、`POST /api/v1/auth/oauth2/test`）；重构公共站点配置为 `/public/settings`（挂载于独立 `PublicController`），支持多身份源元数据聚合发现并保持全量向后兼容；
  - **登录页多 SSO 聚合入口**：登录界面自动检测并聚合展示所有已激活的快捷登录入口（例如 OIDC 图标与 GitHub 闪电图标并列呈现），用户可自由选配登录通道；
  - **安全 Reader 优雅抽象**：重构服务端配置读取体系，引入专用的 `securityReader` 辅助抽象，支持变长参数回退链（Fallback Chain）、类型安全转换与 Scope 解析去重，彻底消除冗余 if 样板代码；
  - **Provider 动态缓存与防脏缓存**：优化 `oidcCache` 全局缓存机制，仅缓存耗时昂贵的远程发现文档与 JWKS 验签器，`oauth2.Config` 始终根据当前请求动态构造，彻底杜绝多域名反代或密钥修改时 `RedirectURL` 脏缓存导致的 `redirect_uri_mismatch` 故障；
  - **Subject 命名空间防碰撞**：自动注入通道命名空间前缀（`oauth2:<id>` 与 `oidc:<sub>`），杜绝不同提供方纯数字 ID 碰撞串登，并对历史存量账号实现无感平滑升级；
  - **端点测试健壮性与超时防护**：通用 OAuth 2.0 测试端点探测增加 5s 超时控制与定制 UA 标识，解决参数自动识别与网络挂起风险。
* **OIDC 回跳与发现文档路径容错 (Fix & OIDC)**：
  - **登录回跳 State 校验修复**：修复携带重定向回跳路径（`redirect` 参数）时 Cookie 中 state 校验失败的缺陷（#194）；
  - **Issuer 路径自动规范化**：自动剔除用户误填的 `/.well-known/openid-configuration` 后缀与多余斜杠，并优化 404 排障指引。
* **移动端小屏响应式排版与视觉细节打磨 (Style & Mobile & UI)**：
  - **单点登录小屏排版与文案精简**：通道选项卡标题精简为 `OIDC` 与 `OAuth 2.0` 并优化间距，彻底消除移动端小屏 `通用 OAu...` 截断省略号；通道主开关条标题精简为 `OIDC 通道` 与 `OAuth 2.0 通道`，状态徽章与标题锁定同行布局防折行；顶部操作栏双按钮等分对称平铺（`grid grid-cols-2`），端点测试按钮全宽单行铺满（`w-full`），回调地址代码框自适应拉伸；
  - **数据仪表盘状态徽章位置优化**：在小屏及窄视口下将数据仪表卡片的服务状态徽章移至右上角固定定位，彻底解决标题行在小屏设备下被挤压换行的排版痛点；
  - **仪表盘文本单行精炼**：优化仪表盘小屏副标题文案，保持单行紧凑对齐；
  - **首页水印与设置中心自适应**：优化首页动态水印在各视口下的防遮挡表现，统一移动端设置面板间距；
  - **UI 视觉图标规范化**：移除设置项提示中突兀的 Emoji 字符，统一替换为系统规范的 Lucide 图标组件（`<HelpCircle>`）。

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

  直接运行附件中的安装程序 `BaihuPanel-Setup-v1.5.5-windows-amd64.exe` 完成安装，程序将自动完成桌面快捷方式创建、开机自启配置与 `127.0.0.1 baihu.local` 域名绑定。安装完成后通过桌面图标或任务栏右下角托盘图标即可直接打开 `http://baihu.local:38052`。

* **方式二：二进制单文件运行（免安装/便携版）**

  解压下载好的 `.zip` 压缩包（如 `baihu-windows-amd64.zip`），进入解压目录并在 PowerShell 中运行：

  ```powershell
  .\baihu.exe server
  ```

---

**访问面板：**
* 启动后访问：`http://baihu.local:38052` 或 `http://localhost:38052` (或单文件默认端口 `http://localhost:8052`)
* **默认账号**：用户名 `admin`，密码见面板首次启动时的控制台日志。
