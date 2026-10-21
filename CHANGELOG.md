# 更新日志 (v1.5.7)

### 2026.10.21 - 任务日志自定义变量提取与通知模板注入、设置中心全局悬浮保存与改动追踪、脚本编辑终端重构及体验加固

🎉 **新增功能与体验升级**
* **任务日志自定义变量提取与通知模板注入 (Major Feature & Notify & Tasks)**：
  - **双轨提取引擎**：
    - **标准协议直出**：任务脚本可直接在标准输出打印 `[BHP_VAR] key=val`、GitHub Actions 兼容语法 `echo "::set-output name=key::val"` 或 `[BHP_JSON] {"key": "val"}`，自动完成结构化变量捕获；
    - **零侵入正则捕获**：无需修改原有脚本代码，在任务通知配置中配置自定义正则表达式（支持命名捕获组 `(?P<name>pattern)` 与单捕获组），从运行日志中无感提取关键指标（如订单号、签到积分、执行状态等）；
  - **五层变量叠加覆盖与安全防篡改**：按 `系统内置保留变量 < 协议直出变量 < 正则捕获变量 < 任务自身参数` 优先级自动合并注入通知模板占位符（如 `{{order_id}}`）；内置系统字段防篡改机制，保障通知模板稳定性；
  - **可视化配置与官方指南**：任务通知抽屉新增「变量正则提取」配置卡片，支持多条规则动态增删与格式验证；官方文档中心完善发布《任务日志变量提取与通知模板注入使用指南》（`docs/guide/notify.md`），提供 Shell/Python/Node.js/Go 全语言示例（#196）。
* **系统设置中心全局悬浮保存栏与改动追踪 (Settings & Web & UI)**：
  - **动态改动追踪器 (Settings Tracker)**：引入 `useSettingsTracker` 组合式状态追踪，深度对比各配置项初始状态，修改时实时在字段旁渲染高对比度「已修改」徽章标记（`ModifiedBadge.vue`）；
  - **全局悬浮保存栏 (FloatingSaveBar)**：站点设置、调度设置与单点登录页面长滚动时，自动在屏幕底部浮出毛玻璃悬浮操作栏，实时统计未保存改动计数，支持一键重置放弃修改或原子落库保存；
  - **通道测试布局调优**：优化各通知通道测试按钮与输入框的间距布局，提升表单视觉均衡性与操作容错率。
* **脚本代码编辑与底部终端面板体验重构 (Editor & Web & Fix)**：
  - **消除参照系内边距导致的巨大空白**：彻底移除容器错误的 `pb-[45%]`（原百分比 padding 按父容器宽度计算导致 298px 巨大空白挤压编辑器），底部控制台改为 Flex 布局内 `shrink-0`，高度统一由 flex 自适应分配（#197）；
  - **动画位移改为流式高度平滑过渡**：面板展开/收回由 translate 位移改为高度过渡（`h-0!` 到 `h-[45%]`），确保过渡全程面板高度与编辑器高度总和严格等于视口容器，消除背景裸露与抽搐闪烁；
  - **Monaco 1px 亚像素缝隙修复与重排兜底**：容器补充 `bg-[#1e1e1e]` 底色，解决 Monaco 高度取整在浅色背景下漏白线的瑕疵；引入 `ResizeObserver` 兜底监听，防止控制台最大化关闭后编辑器未触发重排。
* **系统全局资源隔离与环境变量/机密查询修复 (Env & Security & Fix)**：
  - **解除机密与环境变量 user_id 过滤**：移除环境变量与机密查询中针对 `user_id` 的限制，使其作为系统全局统一资源调度注入，彻底解决后台 Cron 及应用子任务在部分多用户上下文下无法正确获取全局机密与变量的问题。
* **登录日志 IP 详情弹窗排版与字体优化 (Logs & Web & Style)**：
  - **超长文本排版防溢出与截断**：优化国际运营商/组织超长文本展示，引入优雅截断与悬停 Tooltip 完整信息预览，解决弹窗在不同分辨率下的横向溢出与图标排版漂移；
  - **全站 Inter 字体规范与国旗 Emoji 保留**：IP 详情统一西文字体为 Inter，保证数字与字母等宽对齐；调优字体回退栈，保证各国国旗 Emoji 渲染不被降级丢失。
* **CI 流水线历史日志清理保护 (DevOps & CI)**：
  - **Actions 运行记录安全阈值**：优化 GitHub Actions 自动清理工作流，强制保留最近 100 条构建与部署日志，便于历史追踪与故障排查。

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

  直接运行附件中的安装程序 `BaihuPanel-Setup-v1.5.7-windows-amd64.exe` 完成安装，程序将自动完成桌面快捷方式创建、开机自启配置与 `127.0.0.1 baihu.local` 域名绑定。安装完成后通过桌面图标或任务栏右下角托盘图标即可直接打开 `http://baihu.local:38052`。

* **方式二：二进制单文件运行（免安装/便携版）**

  解压下载好的 `.zip` 压缩包（如 `baihu-windows-amd64.zip`），进入解压目录并在 PowerShell 中运行：

  ```powershell
  .\baihu.exe server
  ```

---

**访问面板：**
* 启动后访问：`http://baihu.local:38052` 或 `http://localhost:38052` (或单文件默认端口 `http://localhost:8052`)
* **默认账号**：用户名 `admin`，密码见面板首次启动时的控制台日志。
