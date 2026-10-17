#!/bin/sh
set -e

# ============================================================
# 白虎面板 (Baihu Panel) 容器入口公共函数库 (Entrypoint Common)
# 参考 build-go.sh 日志样式与步骤规范
# ============================================================

# ANSI 颜色与文本样式
C_RESET="\033[0m"
C_BOLD="\033[1m"
C_CYAN="\033[1;36m"
C_GREEN="\033[1;32m"
C_YELLOW="\033[1;33m"
C_BLUE="\033[1;34m"

# 标准日志函数
log_info() {
  printf "${C_CYAN}[Entrypoint]${C_RESET} %s\n" "$*"
}

log_step() {
  printf "\n${C_BLUE}==>${C_RESET} ${C_BOLD}%s${C_RESET}\n" "$*"
}

log_success() {
  printf "${C_GREEN}✔ %s${C_RESET}\n" "$*"
}

log_warn() {
  printf "${C_YELLOW}⚠ %s${C_RESET}\n" "$*"
}

# 兼容原有的 log 函数
log() {
  log_info "$*"
}

# ============================================================
# 1. 版本与系统信息提取
# ============================================================
init_panel_version() {
  PANEL_VER=$(cat /build-info/version.txt 2>/dev/null || true)
  BUILD_TIME=$(cat /build-info/build_time.txt 2>/dev/null || true)

  if [ -z "$PANEL_VER" ] && [ -x /app/baihu ]; then
    PANEL_VER=$(/app/baihu version 2>/dev/null | awk '{print $2}' || true)
  fi
  if [ -z "$BUILD_TIME" ] && [ -x /app/baihu ]; then
    BUILD_TIME=$(/app/baihu version 2>/dev/null | awk -F'Build time: ' '{print $2}' | tr -d ')' || true)
  fi

  PANEL_VER="${PANEL_VER:-unknown}"
  BUILD_TIME="${BUILD_TIME:-unknown}"
  SYS_ARCH="$(uname -m 2>/dev/null || echo 'unknown')"
  SYS_OS="$(uname -s 2>/dev/null || echo 'Linux')"
}

# 打印横幅 Banner
show_banner() {
  _mode_title="${1:-白虎面板 (Baihu Panel) 容器运行环境}"
  init_panel_version

  echo "============================================================"
  printf "          %s\n" "${_mode_title}"
  echo "============================================================"
  log_info "面板程序版本 : ${PANEL_VER}"
  log_info "构建时间戳   : ${BUILD_TIME}"
  log_info "系统平台架构 : ${SYS_OS} / ${SYS_ARCH}"
  echo "============================================================"
}

# ============================================================
# 2. 初始化数据与脚本工作目录
# ============================================================
init_directories() {
  log_step "初始化数据与脚本工作目录"
  mkdir -p \
    /app/data \
    /app/data/scripts \
    /app/configs \
    /app/envs

  if [ -d "/app/example" ]; then
    mkdir -p /app/data/scripts/example
    rsync -a --ignore-existing /app/example/ /app/data/scripts/example/ || true
    log_success "示例脚本已同步至 /app/data/scripts/example"
  else
    log_info "未检测到示例脚本目录，跳过同步"
  fi
}

# ============================================================
# 3. 检查并同步 Mise 基础运行底座
# ============================================================
sync_mise_environment() {
  _base_dir="${1:-/opt/mise-base}"
  _chown_opt="${2:-}"
  _target_dir="${3:-/app/envs/mise}"

  log_step "检查并同步 Mise 基础运行环境"
  mkdir -p "$_target_dir"

  if [ -d "$_base_dir" ]; then
    _curr_ver="${PANEL_VER:-$(cat /build-info/version.txt 2>/dev/null || echo '')}"
    _installed_ver=$(cat "${_target_dir}/.installed_version" 2>/dev/null || echo "")

    _need_sync=0
    if [ ! -f "/tmp/.mise_container_synced" ]; then
      _need_sync=1
    elif [ -n "$_curr_ver" ] && [ "$_curr_ver" != "unknown" ] && [ "$_curr_ver" != "$_installed_ver" ]; then
      _need_sync=1
    elif [ ! -f "${_target_dir}/config.toml" ]; then
      _need_sync=1
    fi

    if [ "$_need_sync" -eq 1 ]; then
      log_info "正在从底座同步 Mise 运行环境 (全新部署或镜像升级)..."
      if [ -n "$_chown_opt" ]; then
        rsync -a --chown="$_chown_opt" --ignore-existing "${_base_dir}/" "${_target_dir}/" || true
      else
        rsync -a --ignore-existing "${_base_dir}/" "${_target_dir}/" || true
      fi
      touch /tmp/.mise_container_synced
      [ -n "$_curr_ver" ] && echo "$_curr_ver" > "${_target_dir}/.installed_version"
      log_success "Mise 基础运行环境同步完成"
    else
      log_success "Mise 基础运行环境已就绪，跳过同步"
    fi
  else
    log_info "未检测到底座环境 (${_base_dir})，跳过同步"
  fi
}

# ============================================================
# 4. 配置运行时环境与系统 PATH
# ============================================================
setup_runtime_env() {
  _mise_dir="${1:-/app/envs/mise}"

  log_step "配置运行时环境与系统 PATH"
  export MISE_DATA_DIR="$_mise_dir"
  export MISE_CONFIG_DIR="$_mise_dir"
  export PATH="${_mise_dir}/shims:${_mise_dir}/bin:$PATH"

  # 默认启用 Python 镜像源
  export PIP_INDEX_URL=${PIP_INDEX_URL:-https://pypi.org/simple}

  # Node 内存限制与全局脚本模块路径
  export NODE_OPTIONS="--max-old-space-size=256"
  export PYTHONPATH=/app/data/scripts:$PYTHONPATH
  log_success "环境变量注入与 PATH 配置完成"
}

# ============================================================
# 5. 验证预装语言与运行时环境
# ============================================================
verify_runtimes() {
  log_step "验证预装语言与运行时环境"
  if [ -x "$(command -v mise)" ]; then
    log_info "  - mise   : $(mise --version 2>/dev/null | head -n 1 || echo "not found")"
  else
    log_info "  - mise   : not installed"
  fi

  if [ -x "$(command -v python)" ]; then
    log_info "  - python : $(python --version 2>&1 | head -n 1) ($(which python 2>/dev/null || true))"
  else
    log_info "  - python : not installed"
  fi

  if [ -x "$(command -v node)" ]; then
    log_info "  - node   : $(node --version 2>&1 | head -n 1) ($(which node 2>/dev/null || true))"
  else
    log_info "  - node   : not installed"
  fi

  if [ -x "$(command -v npm)" ]; then
    log_info "  - npm    : $(npm --version 2>&1 | head -n 1) ($(which npm 2>/dev/null || true))"
  else
    log_info "  - npm    : not installed"
  fi
}

# ============================================================
# 6. 注册全局 CLI 命令与自动补全
# ============================================================
setup_cli_and_completion() {
  _extra_rc="${1:-}"

  log_step "注册全局 CLI 命令与自动补全"
  if [ -x "/app/baihu" ]; then
    ln -sf /app/baihu /usr/local/bin/baihu
  fi

  _rc_targets="/etc/bash.bashrc /etc/bashrc /root/.bashrc"
  [ -n "$_extra_rc" ] && _rc_targets="${_rc_targets} ${_extra_rc}"

  for _rcfile in $_rc_targets; do
    if [ -f "$_rcfile" ] || [ -d "$(dirname "$_rcfile" 2>/dev/null)" ]; then
      if ! grep -q "baihu completion" "$_rcfile" 2>/dev/null; then
        echo 'eval "$(baihu completion bash 2>/dev/null)"' >> "$_rcfile" 2>/dev/null || true
      fi
    fi
  done
  log_success "全局命令 baihu 及 Shell 补全配置就绪"
}

# ============================================================
# 7. 启动主服务
# ============================================================
start_baihu_server() {
  log_step "环境初始化完成，正在启动白虎面板服务..."
  printf "\n${C_GREEN}${C_BOLD}>>> Environment setup complete. Starting Baihu Server...${C_RESET}\n\n"

  cd /app
  # tini 作为 1 号进程回收孤儿进程，避免僵尸进程累积
  exec tini -- baihu server
}
