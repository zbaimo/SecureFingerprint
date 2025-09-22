# 🚀 GitHub仓库设置和自动构建指南

## 📝 第一步：创建GitHub仓库

### 1. 访问GitHub网站
- 打开 [GitHub.com](https://github.com)
- 登录您的账号

### 2. 创建新仓库
- 点击右上角的 "+" 按钮
- 选择 "New repository"
- 仓库名称：`firewall-controller`
- 描述：`智能访问控制系统 - 基于用户指纹的行为分析和风险控制`
- 设置为 Public（这样GitHub Actions免费）
- 不要添加 README、.gitignore 或 license（我们已经有了）
- 点击 "Create repository"

## 🔗 第二步：连接本地仓库到GitHub

创建仓库后，GitHub会显示连接命令。在您的项目目录中执行：

```bash
# 添加远程仓库（替换为您的GitHub用户名）
git remote add origin https://github.com/YOUR_USERNAME/firewall-controller.git

# 推送到GitHub
git branch -M main
git push -u origin main
```

## 🔐 第三步：配置Docker Hub自动构建

### 1. 获取Docker Hub访问令牌
- 登录 [Docker Hub](https://hub.docker.com)
- 点击右上角头像 → Account Settings
- 点击 Security → New Access Token
- 输入描述：`GitHub Actions - Firewall Controller`
- 权限选择：Read, Write, Delete
- 复制生成的令牌（只显示一次！）

### 2. 在GitHub仓库中配置Secrets
- 进入您的GitHub仓库
- 点击 Settings → Secrets and variables → Actions
- 点击 "New repository secret"
- 添加以下Secrets：

| Name | Value |
|------|-------|
| `DOCKER_USERNAME` | 您的Docker Hub用户名 |
| `DOCKER_PASSWORD` | 刚才创建的访问令牌 |

### 3. 触发自动构建
推送代码后，GitHub Actions会自动：
- 构建多平台Docker镜像
- 推送到Docker Hub
- 运行安全扫描
- 生成部署文件

## 📋 准备好的命令

假设您的GitHub用户名是 `zbaimo`，执行以下命令：

```bash
# 1. 添加远程仓库
git remote add origin https://github.com/zbaimo/firewall-controller.git

# 2. 推送代码
git branch -M main
git push -u origin main
```

## 🎯 自动构建结果

构建完成后，您将获得：

### Docker镜像
- `zbaimo/firewall-controller:latest`
- `zbaimo/firewall-controller:v1.0.0`

### 支持的平台
- `linux/amd64` (Intel/AMD 64位)
- `linux/arm64` (ARM 64位，如Apple M1/M2)

### 用户使用方式
```bash
# 快速启动
docker run -d -p 8080:8080 zbaimo/firewall-controller:latest

# 完整部署
curl -O https://raw.githubusercontent.com/zbaimo/firewall-controller/main/docker-compose.prod.yml
docker-compose -f docker-compose.prod.yml up -d
```

## 🔍 监控构建状态

1. 推送代码后，访问仓库的 "Actions" 标签页
2. 查看 "Build and Push Docker Image" 工作流
3. 构建完成后，检查Docker Hub上的镜像

## ⚠️ 注意事项

1. **仓库必须是Public** - 否则GitHub Actions有使用限制
2. **Docker Hub用户名** - 确保在Secrets中配置正确
3. **访问令牌权限** - 确保有写入权限
4. **首次构建** - 可能需要10-15分钟（包含前端构建）

## 📞 如果遇到问题

1. **构建失败** - 检查GitHub Actions日志
2. **推送失败** - 验证Docker Hub凭据
3. **权限问题** - 确认访问令牌权限

准备好了吗？请告诉我您的GitHub用户名，我帮您执行推送命令！
