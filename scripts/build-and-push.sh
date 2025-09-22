#!/bin/bash

# Docker镜像构建和推送脚本
# 用于构建防火墙控制器镜像并推送到Docker Hub

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
DOCKER_USERNAME="${DOCKER_USERNAME:-zbaimo}"
IMAGE_NAME="${IMAGE_NAME:-securefingerprint}"
VERSION="${VERSION:-v1.0.0}"
PLATFORM="${PLATFORM:-linux/amd64,linux/arm64}"

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

# 检查依赖
check_dependencies() {
    log_info "检查构建依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi
    
    if ! docker buildx version &> /dev/null; then
        log_error "Docker Buildx 未安装"
        exit 1
    fi
    
    log_info "依赖检查完成"
}

# 登录Docker Hub
docker_login() {
    log_info "登录Docker Hub..."
    
    if [ -n "$DOCKER_PASSWORD" ]; then
        echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
    else
        docker login -u "$DOCKER_USERNAME"
    fi
    
    log_info "Docker Hub登录成功"
}

# 设置buildx
setup_buildx() {
    log_info "设置Docker Buildx..."
    
    # 创建新的builder实例
    docker buildx create --name multiarch --driver docker-container --use || true
    docker buildx inspect --bootstrap
    
    log_info "Buildx设置完成"
}

# 构建镜像
build_image() {
    log_info "构建Docker镜像..."
    
    # 获取构建信息
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    
    # 镜像标签
    FULL_IMAGE_NAME="$DOCKER_USERNAME/$IMAGE_NAME"
    
    log_info "镜像名称: $FULL_IMAGE_NAME"
    log_info "版本标签: $VERSION, latest"
    log_info "构建平台: $PLATFORM"
    log_info "构建时间: $BUILD_TIME"
    log_info "Git提交: $GIT_COMMIT"
    
    # 构建多平台镜像
    docker buildx build \
        --platform "$PLATFORM" \
        --build-arg VERSION="$VERSION" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        --build-arg GIT_COMMIT="$GIT_COMMIT" \
        --tag "$FULL_IMAGE_NAME:$VERSION" \
        --tag "$FULL_IMAGE_NAME:latest" \
        --push \
        .
    
    log_info "镜像构建和推送完成"
}

# 验证镜像
verify_image() {
    log_info "验证镜像..."
    
    FULL_IMAGE_NAME="$DOCKER_USERNAME/$IMAGE_NAME"
    
    # 拉取并测试镜像
    docker pull "$FULL_IMAGE_NAME:latest"
    
    # 运行健康检查
    log_info "运行镜像健康检查..."
    docker run --rm "$FULL_IMAGE_NAME:latest" --version || true
    
    log_info "镜像验证完成"
}

# 生成使用说明
generate_usage() {
    cat << EOF

========================================
🛡️  防火墙控制器镜像构建完成
========================================

📦 镜像信息:
  - 名称: $DOCKER_USERNAME/$IMAGE_NAME
  - 版本: $VERSION, latest
  - 平台: $PLATFORM

🚀 使用方法:

1. 快速启动:
   docker run -d -p 8080:8080 $DOCKER_USERNAME/$IMAGE_NAME:latest

2. 使用Docker Compose:
   # 下载配置文件
   curl -O https://raw.githubusercontent.com/your-org/firewall-controller/main/docker-compose.yml
   
   # 启动服务
   docker-compose up -d

3. 环境变量配置:
   docker run -d \\
     -p 8080:8080 \\
     -e REDIS_HOST=redis \\
     -e MYSQL_HOST=mysql \\
     $DOCKER_USERNAME/$IMAGE_NAME:latest

4. 数据持久化:
   docker run -d \\
     -p 8080:8080 \\
     -v /host/logs:/app/logs \\
     -v /host/data:/app/data \\
     $DOCKER_USERNAME/$IMAGE_NAME:latest

📚 更多信息:
  - 文档: https://github.com/your-org/firewall-controller
  - Docker Hub: https://hub.docker.com/r/$DOCKER_USERNAME/$IMAGE_NAME

========================================

EOF
}

# 清理
cleanup() {
    log_info "清理构建环境..."
    docker buildx rm multiarch || true
}

# 主函数
main() {
    echo "🐳 开始构建防火墙控制器Docker镜像..."
    echo
    
    # 检查参数
    if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
        echo "用法: $0 [选项]"
        echo
        echo "选项:"
        echo "  --version VERSION     设置镜像版本 (默认: v1.0.0)"
        echo "  --username USERNAME   设置Docker Hub用户名 (默认: zbaimo)"
        echo "  --platform PLATFORM   设置构建平台 (默认: linux/amd64,linux/arm64)"
        echo "  --no-push            只构建不推送"
        echo "  --help               显示帮助信息"
        echo
        echo "环境变量:"
        echo "  DOCKER_USERNAME      Docker Hub用户名"
        echo "  DOCKER_PASSWORD      Docker Hub密码"
        echo "  VERSION              镜像版本"
        echo
        exit 0
    fi
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --username)
                DOCKER_USERNAME="$2"
                shift 2
                ;;
            --platform)
                PLATFORM="$2"
                shift 2
                ;;
            --no-push)
                NO_PUSH=true
                shift
                ;;
            *)
                log_error "未知参数: $1"
                exit 1
                ;;
        esac
    done
    
    # 检查依赖
    check_dependencies
    
    # 登录Docker Hub（如果需要推送）
    if [ "$NO_PUSH" != "true" ]; then
        docker_login
    fi
    
    # 设置buildx
    setup_buildx
    
    # 构建镜像
    if [ "$NO_PUSH" = "true" ]; then
        log_info "只构建镜像，不推送..."
        # 修改构建命令，移除--push参数
        BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
        GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        FULL_IMAGE_NAME="$DOCKER_USERNAME/$IMAGE_NAME"
        
        docker buildx build \
            --platform "$PLATFORM" \
            --build-arg VERSION="$VERSION" \
            --build-arg BUILD_TIME="$BUILD_TIME" \
            --build-arg GIT_COMMIT="$GIT_COMMIT" \
            --tag "$FULL_IMAGE_NAME:$VERSION" \
            --tag "$FULL_IMAGE_NAME:latest" \
            --load \
            .
    else
        build_image
        verify_image
    fi
    
    # 生成使用说明
    generate_usage
    
    # 清理
    cleanup
    
    log_info "🎉 构建完成！"
}

# 错误处理
trap 'log_error "构建过程中出现错误"; cleanup; exit 1' ERR

# 执行主函数
main "$@"
