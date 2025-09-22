# 🐳 Docker镜像构建和上传指南

由于您的本地环境没有Docker，这里提供几种方式来构建和上传镜像：

## 🛠️ 方法一：安装Docker Desktop (推荐)

### 1. 下载并安装Docker Desktop
- 访问 [Docker Desktop官网](https://www.docker.com/products/docker-desktop/)
- 下载Windows版本并安装
- 启动Docker Desktop

### 2. 构建和推送镜像
```bash
# 在项目根目录执行
cd C:\Users\ZBaimo\Desktop\SecureFingerprint

# 登录Docker Hub
docker login

# 构建镜像
docker build -t zbaimo/firewall-controller:latest -t zbaimo/firewall-controller:v1.0.0 .

# 推送镜像
docker push zbaimo/firewall-controller:latest
docker push zbaimo/firewall-controller:v1.0.0
```

## 🛠️ 方法二：使用GitHub Actions (自动化)

### 1. 创建GitHub仓库
```bash
# 初始化Git仓库
git init
git add .
git commit -m "Initial commit: Firewall Controller v1.0.0"

# 添加远程仓库
git remote add origin https://github.com/zbaimo/firewall-controller.git
git push -u origin main
```

### 2. 配置GitHub Secrets
在GitHub仓库设置中添加以下Secrets：
- `DOCKER_USERNAME`: 您的Docker Hub用户名
- `DOCKER_PASSWORD`: 您的Docker Hub密码或访问令牌

### 3. 自动构建
推送代码到GitHub后，GitHub Actions会自动构建并推送镜像到Docker Hub。

## 🛠️ 方法三：使用云端构建服务

### 1. Docker Hub自动构建
- 连接GitHub仓库到Docker Hub
- 配置自动构建规则
- 推送代码即可触发构建

### 2. 其他云服务
- **Google Cloud Build**
- **AWS CodeBuild**
- **Azure Container Registry**

## 📦 构建后的镜像信息

构建完成后，您的镜像将包含：

### 镜像标签
- `zbaimo/firewall-controller:latest` - 最新版本
- `zbaimo/firewall-controller:v1.0.0` - 特定版本

### 镜像特性
- **多平台支持**: linux/amd64, linux/arm64
- **体积优化**: 使用Alpine Linux基础镜像
- **安全性**: 非root用户运行
- **健康检查**: 内置健康检查机制
- **配置灵活**: 支持环境变量配置

### 使用方法
```bash
# 快速启动
docker run -d -p 8080:8080 zbaimo/firewall-controller:latest

# 完整部署
docker-compose -f docker-compose.prod.yml up -d
```

## 🔧 手动构建命令

如果您有Docker环境，可以使用以下命令：

```bash
# 基础构建
docker build -t zbaimo/firewall-controller:latest .

# 多平台构建
docker buildx create --use
docker buildx build --platform linux/amd64,linux/arm64 -t zbaimo/firewall-controller:latest --push .

# 带版本信息的构建
docker build \
  --build-arg VERSION=v1.0.0 \
  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  -t zbaimo/firewall-controller:v1.0.0 \
  -t zbaimo/firewall-controller:latest \
  .

# 推送到Docker Hub
docker push zbaimo/firewall-controller:v1.0.0
docker push zbaimo/firewall-controller:latest
```

## 🚀 部署使用

构建完成后，用户可以通过以下方式使用：

```bash
# 1. 直接运行
docker run -d \
  --name firewall-controller \
  -p 8080:8080 \
  -e REDIS_HOST=your-redis-host \
  -e MYSQL_HOST=your-mysql-host \
  zbaimo/firewall-controller:latest

# 2. 使用Docker Compose
curl -O https://raw.githubusercontent.com/zbaimo/firewall-controller/main/docker-compose.prod.yml
docker-compose -f docker-compose.prod.yml up -d

# 3. Kubernetes部署
kubectl apply -f https://raw.githubusercontent.com/zbaimo/firewall-controller/main/k8s/deployment.yaml
```

## 📋 下一步操作

1. **安装Docker Desktop** 或选择云端构建方案
2. **创建Docker Hub账号** (如果还没有)
3. **执行构建命令** 或设置自动构建
4. **测试镜像** 确保功能正常
5. **更新文档** 添加使用说明

选择最适合您的方法来构建和上传镜像！
