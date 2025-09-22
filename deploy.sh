#!/bin/bash

# SecureFingerprint 快速部署脚本

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

# 显示帮助信息
show_help() {
    cat << EOF
🛡️ SecureFingerprint 部署脚本

用法: $0 [模式] [选项]

部署模式:
  basic      基础模式 (仅核心应用)
  full       完整模式 (包含数据库)
  monitoring 监控模式 (包含监控系统)
  proxy      代理模式 (包含Nginx)
  all        全功能模式 (包含所有服务)

选项:
  --pull     拉取最新镜像
  --build    本地构建镜像
  --down     停止并删除服务
  --logs     查看日志
  --status   查看服务状态
  --help     显示帮助信息

示例:
  $0 basic                    # 启动基础版本
  $0 full --pull             # 拉取最新镜像并启动完整版本
  $0 --down                  # 停止所有服务
  $0 --logs securefingerprint # 查看应用日志

EOF
}

# 检查Docker环境
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi

    if ! docker compose version &> /dev/null; then
        log_error "Docker Compose 未安装"
        exit 1
    fi

    log_info "Docker 环境检查完成"
}

# 拉取镜像
pull_images() {
    log_info "拉取最新镜像..."
    docker compose pull
}

# 创建必要目录
create_directories() {
    log_info "创建必要目录..."
    mkdir -p mysql_data nginx_logs nginx
    chmod 755 mysql_data nginx_logs
}

# 部署服务
deploy_service() {
    local mode=$1
    local pull_flag=$2
    local build_flag=$3

    create_directories

    if [ "$pull_flag" = "true" ]; then
        pull_images
    fi

    case $mode in
        "basic")
            log_info "启动基础模式..."
            docker compose up -d securefingerprint
            ;;
        "full")
            log_info "启动完整模式..."
            docker compose --profile full up -d
            ;;
        "monitoring")
            log_info "启动监控模式..."
            docker compose --profile monitoring up -d
            ;;
        "proxy")
            log_info "启动代理模式..."
            docker compose --profile proxy up -d
            ;;
        "all")
            log_info "启动全功能模式..."
            docker compose --profile full --profile monitoring --profile proxy up -d
            ;;
        *)
            log_error "无效的部署模式: $mode"
            show_help
            exit 1
            ;;
    esac
}

# 停止服务
stop_services() {
    log_info "停止所有服务..."
    docker compose down
    log_info "服务已停止"
}

# 查看日志
show_logs() {
    local service=$1
    if [ -n "$service" ]; then
        docker compose logs -f "$service"
    else
        docker compose logs -f
    fi
}

# 查看状态
show_status() {
    log_info "服务状态:"
    docker compose ps
    
    echo
    log_info "网络状态:"
    docker network ls | grep secure
    
    echo
    log_info "数据卷状态:"
    docker volume ls | grep secure
}

# 显示访问信息
show_access_info() {
    echo
    echo "=========================================="
    echo "🛡️ SecureFingerprint 部署完成"
    echo "=========================================="
    echo
    echo "📊 主应用: http://localhost:8080"
    echo "💚 健康检查: http://localhost:8080/api/v1/system/health"
    echo "📋 系统信息: http://localhost:8080/api/v1/system/info"
    echo
    
    # 检查可选服务
    if docker compose ps | grep -q secure-mysql; then
        echo "🗄️ MySQL: localhost:3306"
    fi
    
    if docker compose ps | grep -q secure-redis; then
        echo "📦 Redis: localhost:6379"
    fi
    
    if docker compose ps | grep -q secure-nginx; then
        echo "🌐 Nginx: http://localhost"
    fi
    
    if docker compose ps | grep -q secure-grafana; then
        echo "📈 Grafana: http://localhost:3000 (admin/admin123)"
    fi
    
    if docker compose ps | grep -q secure-prometheus; then
        echo "📊 Prometheus: http://localhost:9090"
    fi
    
    echo
    echo "📋 常用命令:"
    echo "  查看日志: $0 --logs"
    echo "  查看状态: $0 --status"
    echo "  停止服务: $0 --down"
    echo "=========================================="
}

# 主函数
main() {
    local mode=""
    local pull_flag=false
    local build_flag=false
    local down_flag=false
    local logs_flag=false
    local status_flag=false
    local logs_service=""

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            basic|full|monitoring|proxy|all)
                mode=$1
                shift
                ;;
            --pull)
                pull_flag=true
                shift
                ;;
            --build)
                build_flag=true
                shift
                ;;
            --down)
                down_flag=true
                shift
                ;;
            --logs)
                logs_flag=true
                logs_service=$2
                if [[ $2 && ! $2 =~ ^-- ]]; then
                    shift
                fi
                shift
                ;;
            --status)
                status_flag=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # 检查Docker环境
    check_docker

    # 执行操作
    if [ "$down_flag" = "true" ]; then
        stop_services
    elif [ "$logs_flag" = "true" ]; then
        show_logs "$logs_service"
    elif [ "$status_flag" = "true" ]; then
        show_status
    elif [ -n "$mode" ]; then
        deploy_service "$mode" "$pull_flag" "$build_flag"
        show_access_info
    else
        log_info "启动基础模式..."
        deploy_service "basic" "$pull_flag" "$build_flag"
        show_access_info
    fi
}

# 错误处理
trap 'log_error "部署过程中出现错误"; exit 1' ERR

# 执行主函数
main "$@"
