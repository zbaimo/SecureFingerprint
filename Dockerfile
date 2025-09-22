# 多阶段构建 - 前端构建阶段
FROM node:18-alpine AS frontend-builder

# 设置工作目录
WORKDIR /app/webui

# 复制前端依赖文件
COPY webui/package*.json ./

# 安装依赖
RUN npm ci --only=production

# 复制前端源码
COPY webui/ ./

# 构建前端
RUN npm run build

# Go构建阶段
FROM golang:1.21-alpine AS backend-builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata wget

# 设置工作目录
WORKDIR /app

# 复制Go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源码
COPY . .

# 复制前端构建产物
COPY --from=frontend-builder /app/webui/dist ./webui/build

# 设置构建参数
ARG VERSION=v1.0.0
ARG BUILD_TIME
ARG GIT_COMMIT

# 构建Go应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT} -s -w" \
    -o firewall-controller ./cmd/server

# 最终运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata wget curl

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=backend-builder /app/firewall-controller .

# 复制配置文件
COPY --from=backend-builder /app/configs ./configs

# 复制前端构建产物
COPY --from=backend-builder /app/webui/build ./webui/build

# 创建必要目录
RUN mkdir -p logs data backups && \
    chown -R appuser:appgroup /app

# 复制启动脚本
COPY scripts/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/system/health || exit 1

# 设置标签
LABEL maintainer="Firewall Controller Team" \
      version="1.0.0" \
      description="智能访问控制系统" \
      org.opencontainers.image.title="Firewall Controller" \
      org.opencontainers.image.description="基于用户指纹的智能访问控制系统" \
      org.opencontainers.image.version="1.0.0" \
      org.opencontainers.image.vendor="ZBaimo" \
      org.opencontainers.image.licenses="MIT"

# 启动应用
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["./firewall-controller"]
