#!/bin/bash

# ============================================
# Shiny Collection - 一键启动脚本
# 同时启动 Go 后端和 Vue 前端开发服务器
# ============================================

set -e

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$PROJECT_DIR/backend"
FRONTEND_DIR="$PROJECT_DIR/frontend"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log_info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_cmd()   { echo -e "${CYAN}[CMD]${NC} $1"; }

# 清理函数：退出时杀掉所有子进程
cleanup() {
    echo ""
    log_warn "正在停止服务..."
    if [ -n "$BACKEND_PID" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
        kill "$BACKEND_PID" 2>/dev/null
        log_info "后端服务已停止 (PID: $BACKEND_PID)"
    fi
    if [ -n "$FRONTEND_PID" ] && kill -0 "$FRONTEND_PID" 2>/dev/null; then
        kill "$FRONTEND_PID" 2>/dev/null
        log_info "前端服务已停止 (PID: $FRONTEND_PID)"
    fi
    exit 0
}

trap cleanup SIGINT SIGTERM EXIT

# ========== 环境检查 ==========

log_info "===== Shiny Collection 启动器 ====="
echo ""

# 检查 Go
if ! command -v go &>/dev/null; then
    log_error "未找到 Go，请先安装 Go 1.22+"
    exit 1
fi
log_info "Go: $(go version)"

# 检查 Node.js
if ! command -v node &>/dev/null; then
    log_error "未找到 Node.js，请先安装 Node.js 18+"
    exit 1
fi
log_info "Node: $(node --version)"

# 检查 npm
if ! command -v npm &>/dev/null; then
    log_error "未找到 npm"
    exit 1
fi
log_info "npm: $(npm --version)"
echo ""

# ========== 依赖安装（如需要） ==========

# 检查 Go 依赖
if [ ! -f "$BACKEND_DIR/go.sum" ]; then
    log_info "正在安装 Go 依赖..."
    log_cmd "go mod tidy"
    (cd "$BACKEND_DIR" && go mod tidy)
    log_info "Go 依赖安装完成"
fi

# 检查 Node 依赖
if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
    log_info "正在安装前端依赖..."
    log_cmd "npm install"
    (cd "$FRONTEND_DIR" && npm install)
    log_info "前端依赖安装完成"
fi

# ========== 启动后端 ==========

# 清理残留的后端进程
if lsof -i :8080 &>/dev/null 2>&1; then
    log_warn "端口 8080 被占用，正在清理..."
    fuser -k 8080/tcp 2>/dev/null
    sleep 1
fi

log_info "正在启动后端服务 (端口 8080)..."
(cd "$BACKEND_DIR" && go run cmd/server/main.go) &
BACKEND_PID=$!
echo "  → PID: $BACKEND_PID"

# 等待后端就绪
log_info "等待后端启动..."
for i in $(seq 1 30); do
    if curl -s http://localhost:8080/api/health >/dev/null 2>&1; then
        log_info "后端服务已就绪 ✅"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "后端启动超时，请检查日志"
        exit 1
    fi
    sleep 1
done

echo ""

# ========== 启动前端 ==========

log_info "正在启动前端开发服务器 (端口 5173)..."
(cd "$FRONTEND_DIR" && npm run dev) &
FRONTEND_PID=$!
echo "  → PID: $FRONTEND_PID"

# 等一会儿让前端启动
sleep 2

echo ""
log_info "=================================="
log_info "✨ 所有服务已启动！"
log_info ""
log_info "   🌐 前端地址: ${CYAN}http://localhost:5173${NC}"
log_info "   🔗 后端 API: ${CYAN}http://localhost:8080/api/v1${NC}"
log_info "   📊 健康检查: ${CYAN}http://localhost:8080/api/health${NC}"
log_info ""
log_info "   按 ${YELLOW}Ctrl+C${NC} 停止所有服务"
log_info "=================================="
echo ""

# 前台等待子进程（同时监控两个进程）
while true; do
    if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
        log_error "后端进程意外退出"
        exit 1
    fi
    if ! kill -0 "$FRONTEND_PID" 2>/dev/null; then
        log_error "前端进程意外退出"
        exit 1
    fi
    sleep 2
done
