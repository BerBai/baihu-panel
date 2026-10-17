#!/bin/sh
set -e
export LANG=C.UTF-8
export LC_ALL=C.UTF-8
export MISE_HIDE_UPDATE_WARNING=1

# 加载公共 Entrypoint 库
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMMON_SH="${SCRIPT_DIR}/docker-entrypoint-common.sh"
[ ! -f "$COMMON_SH" ] && COMMON_SH="/app/docker-entrypoint-common.sh"
. "$COMMON_SH"

show_banner "${BAIHU_BANNER_TITLE:-白虎面板 (Baihu Panel) 容器运行环境}"
init_directories
sync_mise_environment "/opt/mise-base"
setup_runtime_env
verify_runtimes
setup_cli_and_completion
start_baihu_server "$@"
