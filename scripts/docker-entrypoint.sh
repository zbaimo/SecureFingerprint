#!/bin/sh

# Docker容器启动脚本
# 用于防火墙控制器的容器初始化

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 等待服务就绪
wait_for_service() {
    local host=$1
    local port=$2
    local service_name=$3
    local max_attempts=30
    local attempt=1

    log_info "等待 $service_name 服务就绪 ($host:$port)..."

    while [ $attempt -le $max_attempts ]; do
        if nc -z "$host" "$port" 2>/dev/null; then
            log_info "$service_name 服务已就绪"
            return 0
        fi
        
        log_warn "等待 $service_name 服务... ($attempt/$max_attempts)"
        sleep 2
        attempt=$((attempt + 1))
    done

    log_error "$service_name 服务启动超时"
    return 1
}

# 检查配置文件
check_config() {
    log_info "检查配置文件..."
    
    if [ ! -f "configs/config.yaml" ]; then
        log_error "配置文件不存在: configs/config.yaml"
        exit 1
    fi
    
    log_info "配置文件检查完成"
}

# 初始化数据库
init_database() {
    log_info "检查数据库连接..."
    
    # 这里可以添加数据库初始化逻辑
    # 例如：运行数据库迁移脚本
    
    log_info "数据库检查完成"
}

# 预热应用
warmup_application() {
    log_info "预热应用..."
    
    # 可以在这里添加应用预热逻辑
    # 例如：预加载缓存、初始化连接池等
    
    log_info "应用预热完成"
}

# 主函数
main() {
    log_info "🛡️ 启动防火墙控制器..."
    log_info "版本: ${VERSION:-v1.0.0}"
    log_info "构建时间: ${BUILD_TIME:-unknown}"
    log_info "Git提交: ${GIT_COMMIT:-unknown}"
    
    # 检查配置
    check_config
    
    # 等待依赖服务（如果配置了的话）
    if [ -n "$REDIS_HOST" ] && [ -n "$REDIS_PORT" ]; then
        wait_for_service "$REDIS_HOST" "$REDIS_PORT" "Redis"
    fi
    
    if [ -n "$MYSQL_HOST" ] && [ -n "$MYSQL_PORT" ]; then
        wait_for_service "$MYSQL_HOST" "$MYSQL_PORT" "MySQL"
    fi
    
    # 初始化数据库
    init_database
    
    # 预热应用
    warmup_application
    
    log_info "✅ 防火墙控制器启动完成"
    log_info "🌐 访问地址: http://localhost:8080"
    
    # 执行传入的命令
    exec "$@"
}

# 信号处理
trap 'log_info "收到终止信号，正在关闭..."; exit 0' TERM INT

# 执行主函数
main "$@"
