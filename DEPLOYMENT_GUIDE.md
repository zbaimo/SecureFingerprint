# 🚀 防火墙控制器部署指南

## 📦 项目概述

这是一个完整的智能访问控制系统，包含：
- Go后端服务
- Vue3前端界面  
- Redis缓存
- MySQL数据库
- Nginx代理配置

## 🎯 快速部署

### 使用Docker Hub镜像（推荐）

一旦镜像上传到Docker Hub，用户可以直接使用：

```bash
# 1. 下载部署配置
curl -O https://raw.githubusercontent.com/zbaimo/SecureFingerprint/main/docker-compose.prod.yml

# 2. 启动服务
docker-compose -f docker-compose.prod.yml up -d

# 3. 访问系统
open http://localhost:8080
```

### 本地构建部署

```bash
# 1. 克隆仓库
git clone https://github.com/zbaimo/SecureFingerprint.git
cd SecureFingerprint

# 2. 构建并启动
make up

# 或者
docker-compose up -d
```

## 🔧 配置说明

### 环境变量
- `MYSQL_ROOT_PASSWORD` - MySQL root密码
- `MYSQL_PASSWORD` - 应用数据库密码
- `REDIS_PASSWORD` - Redis密码（可选）

### 端口映射
- `8080` - 主应用端口
- `3306` - MySQL端口
- `6379` - Redis端口
- `80/443` - Nginx端口

## 📚 功能特性

- ✅ 用户指纹识别
- ✅ 行为分析
- ✅ 风险评估
- ✅ 访问控制
- ✅ 实时监控
- ✅ 管理界面

## 🌐 访问地址

部署完成后：
- **管理后台**: http://localhost:8080
- **API文档**: http://localhost:8080/api/v1
- **健康检查**: http://localhost:8080/health

## 📞 技术支持

如有问题，请查看：
- [完整文档](README.md)
- [配置指南](docs/nginx-proxy-setup.md)
- [故障排除](build-docker.md)
