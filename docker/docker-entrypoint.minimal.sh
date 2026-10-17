#!/bin/sh
set -e
export LANG=C.UTF-8
export LC_ALL=C.UTF-8
export MISE_HIDE_UPDATE_WARNING=1

# 指定 Minimal 模式标识并调用主入口
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
export BAIHU_BANNER_MODE="极简版 (Minimal)"
exec "${SCRIPT_DIR}/docker-entrypoint.sh" "$@"
