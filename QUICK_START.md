# 🚀 SecureFingerprint 快速启动指南

## ⚡ 一键部署

### 方式一：使用部署脚本（推荐）
```bash
# 下载部署脚本
curl -O https://raw.githubusercontent.com/zbaimo/SecureFingerprint/main/deploy.sh
chmod +x deploy.sh

# 快速启动
./deploy.sh basic

# 或启动完整版本
./deploy.sh full --pull
```

### 方式二：直接使用Docker Compose
```bash
# 下载配置文件
curl -O https://raw.githubusercontent.com/zbaimo/SecureFingerprint/main/compose.yaml

# 启动基础版本
docker compose up -d

# 或启动完整版本
docker compose --profile full up -d
```

### 方式三：单容器启动
```bash
# 最简单的启动方式
docker run -d -p 8080:8080 zbaimo/securefingerprint:latest

# 访问系统
open http://localhost:8080
```

## 🎯 部署模式选择

| 模式 | 命令 | 包含服务 | 适用场景 |
|------|------|----------|----------|
| **基础** | `docker compose up -d` | 核心应用 | 快速测试、轻量部署 |
| **完整** | `docker compose --profile full up -d` | 应用+数据库 | 生产环境、数据持久化 |
| **监控** | `docker compose --profile monitoring up -d` | 应用+监控 | 性能监控、运维管理 |
| **代理** | `docker compose --profile proxy up -d` | 应用+Nginx | 反向代理、负载均衡 |
| **全功能** | `docker compose --profile full --profile monitoring --profile proxy up -d` | 所有服务 | 完整生产环境 |

## 🔧 配置自定义

### 环境变量配置
创建 `.env` 文件：
```bash
# 数据库配置
MYSQL_ROOT_PASSWORD=your_secure_root_password
MYSQL_PASSWORD=your_secure_app_password

# 监控配置
GRAFANA_PASSWORD=your_grafana_password

# 时区设置
TZ=Asia/Shanghai
```

### 端口自定义
如果端口冲突，修改 `compose.yaml`：
```yaml
services:
  securefingerprint:
    ports:
      - "8081:8080"  # 改为8081端口
```

## 📊 访问地址

部署完成后的访问地址：

| 服务 | 地址 | 说明 |
|------|------|------|
| **主应用** | http://localhost:8080 | 防火墙控制器主界面 |
| **API文档** | http://localhost:8080/api/v1/system/info | 系统API信息 |
| **健康检查** | http://localhost:8080/api/v1/system/health | 服务健康状态 |
| **Grafana** | http://localhost:3000 | 监控仪表板 (admin/admin123) |
| **Prometheus** | http://localhost:9090 | 指标收集系统 |
| **Nginx** | http://localhost | 反向代理入口 |

## 🛠️ 管理命令

### 查看状态
```bash
# 查看运行状态
docker compose ps

# 查看详细信息
./deploy.sh --status
```

### 查看日志
```bash
# 查看应用日志
docker compose logs securefingerprint

# 实时查看日志
docker compose logs -f securefingerprint

# 查看所有服务日志
./deploy.sh --logs
```

### 重启服务
```bash
# 重启主应用
docker compose restart securefingerprint

# 重启所有服务
docker compose restart
```

### 更新服务
```bash
# 拉取最新镜像并重启
docker compose pull
docker compose up -d

# 或使用脚本
./deploy.sh basic --pull
```

### 停止服务
```bash
# 停止服务
docker compose down

# 停止并删除数据
docker compose down -v

# 或使用脚本
./deploy.sh --down
```

## 🔍 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   # 检查端口占用
   netstat -tlnp | grep :8080
   
   # 修改端口或停止占用服务
   ```

2. **权限问题**
   ```bash
   # 修复目录权限
   sudo chown -R $USER:$USER mysql_data nginx_logs
   ```

3. **镜像拉取失败**
   ```bash
   # 手动拉取镜像
   docker pull zbaimo/securefingerprint:latest
   ```

4. **服务启动失败**
   ```bash
   # 查看详细日志
   docker compose logs securefingerprint
   
   # 检查健康状态
   docker compose ps
   ```

### 完全重置
```bash
# 停止所有服务并清理
docker compose down -v
docker system prune -f

# 重新部署
./deploy.sh basic --pull
```

## 📞 获取帮助

- 查看部署脚本帮助：`./deploy.sh --help`
- 检查GitHub仓库：https://github.com/zbaimo/SecureFingerprint
- 查看Docker Hub：https://hub.docker.com/r/zbaimo/securefingerprint

---

🎉 **现在您可以快速部署SecureFingerprint了！**
