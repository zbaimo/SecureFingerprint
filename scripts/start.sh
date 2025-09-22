#!/bin/bash

# 防火墙控制器启动脚本
# Author: Firewall Controller Team
# Version: 1.0.0

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查系统依赖..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    # 检查Make
    if ! command -v make &> /dev/null; then
        log_warn "Make 未安装，部分功能可能无法使用"
    fi
    
    log_info "依赖检查完成"
}

# 检查端口占用
check_ports() {
    log_info "检查端口占用..."
    
    ports=(80 3306 6379 8080)
    
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            log_warn "端口 $port 已被占用"
            read -p "是否继续？(y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                log_error "启动取消"
                exit 1
            fi
        fi
    done
    
    log_info "端口检查完成"
}

# 创建必要目录
create_directories() {
    log_info "创建必要目录..."
    
    mkdir -p logs
    mkdir -p backups
    mkdir -p docker/nginx/ssl
    mkdir -p docker/grafana/provisioning
    mkdir -p docker/prometheus
    
    log_info "目录创建完成"
}

# 生成配置文件
generate_config() {
    log_info "生成配置文件..."
    
    # 如果配置文件不存在，从模板复制
    if [ ! -f "docker/app/config.yaml" ]; then
        log_warn "配置文件不存在，使用默认配置"
        cp configs/config.yaml docker/app/config.yaml
    fi
    
    log_info "配置文件准备完成"
}

# 构建镜像
build_image() {
    log_info "构建Docker镜像..."
    
    if command -v make &> /dev/null; then
        make docker-build
    else
        docker build -t firewall-controller:v1.0.0 .
    fi
    
    log_info "镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    # 先启动基础服务
    log_info "启动基础服务 (Redis, MySQL)..."
    docker-compose up -d redis mysql
    
    # 等待服务启动
    log_info "等待数据库服务启动..."
    sleep 15
    
    # 启动应用服务
    log_info "启动应用服务..."
    docker-compose up -d firewall-controller
    
    # 启动Nginx
    log_info "启动Nginx..."
    docker-compose up -d nginx
    
    log_info "所有服务启动完成"
}

# 检查服务状态
check_services() {
    log_info "检查服务状态..."
    
    services=("redis" "mysql" "firewall-controller" "nginx")
    
    for service in "${services[@]}"; do
        if docker-compose ps $service | grep -q "Up"; then
            log_info "✓ $service 运行正常"
        else
            log_error "✗ $service 启动失败"
        fi
    done
}

# 显示访问信息
show_access_info() {
    echo
    echo "=========================================="
    echo "🛡️  防火墙控制器启动完成"
    echo "=========================================="
    echo
    echo "📊 管理后台: http://localhost"
    echo "🔧 API文档: http://localhost/api/v1"
    echo "💾 数据库: localhost:3306"
    echo "🗄️  缓存: localhost:6379"
    echo
    echo "📋 常用命令:"
    echo "  查看日志: make logs"
    echo "  停止服务: make down"
    echo "  重启服务: make restart"
    echo
    echo "📚 更多信息请查看 README.md"
    echo "=========================================="
}

# 主函数
main() {
    echo "🚀 启动防火墙控制器..."
    echo
    
    # 检查依赖
    check_dependencies
    
    # 检查端口
    check_ports
    
    # 创建目录
    create_directories
    
    # 生成配置
    generate_config
    
    # 构建镜像
    build_image
    
    # 启动服务
    start_services
    
    # 检查状态
    sleep 5
    check_services
    
    # 显示访问信息
    show_access_info
}

# 处理中断信号
trap 'log_error "启动被中断"; exit 1' INT TERM

# 执行主函数
main "$@"
