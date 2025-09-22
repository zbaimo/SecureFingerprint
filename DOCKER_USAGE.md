# 🐳 Docker Compose 使用指南

## 🚀 快速启动

### 基础版本（推荐）
```bash
# 仅启动核心应用
docker-compose up -d

# 访问系统
open http://localhost:8080
```

### 完整版本（包含数据库）
```bash
# 启动完整功能（包含Redis和MySQL）
docker-compose --profile full up -d

# 查看状态
docker-compose ps
```

### 监控版本（包含监控系统）
```bash
# 启动监控功能（包含Prometheus和Grafana）
docker-compose --profile monitoring up -d

# 访问监控
open http://localhost:3000  # Grafana (admin/admin123)
open http://localhost:9090  # Prometheus
```

### 代理版本（包含Nginx）
```bash
# 启动代理功能
docker-compose --profile proxy up -d

# 通过Nginx访问
open http://localhost  # 通过80端口访问
```

### 全功能版本
```bash
# 启动所有服务
docker-compose --profile full --profile monitoring --profile proxy up -d
```

## 🔧 配置选项

### 环境变量
创建 `.env` 文件来自定义配置：

```bash
# 数据库密码
MYSQL_ROOT_PASSWORD=your_secure_root_password
MYSQL_PASSWORD=your_secure_password

# Grafana密码
GRAFANA_PASSWORD=your_grafana_password

# 时区设置
TZ=Asia/Shanghai
```

### 端口映射
- `8080` - 主应用端口
- `6379` - Redis端口（可选）
- `3306` - MySQL端口（可选）
- `80/443` - Nginx端口（可选）
- `9090` - Prometheus端口（可选）
- `3000` - Grafana端口（可选）

## 📋 常用命令

### 启动和停止
```bash
# 启动基础版本
docker-compose up -d

# 启动完整版本
docker-compose --profile full up -d

# 停止所有服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v
```

### 查看状态
```bash
# 查看运行状态
docker-compose ps

# 查看日志
docker-compose logs securefingerprint

# 实时查看日志
docker-compose logs -f securefingerprint

# 查看所有服务日志
docker-compose logs -f
```

### 维护操作
```bash
# 重启服务
docker-compose restart securefingerprint

# 重新构建镜像
docker-compose build --no-cache

# 更新镜像
docker-compose pull
docker-compose up -d
```

## 🔍 故障排除

### 端口冲突
如果端口被占用，修改docker-compose.yml中的端口映射：
```yaml
ports:
  - "8081:8080"  # 改为8081端口
```

### 权限问题
```bash
# 给日志目录设置权限
mkdir -p logs
chmod 755 logs
```

### 网络问题
```bash
# 重建网络
docker-compose down
docker network prune
docker-compose up -d
```

### 查看详细错误
```bash
# 查看容器详细信息
docker inspect securefingerprint

# 进入容器调试
docker-compose exec securefingerprint sh
```

## 📊 监控和管理

### Grafana仪表板
- 访问：http://localhost:3000
- 用户名：admin
- 密码：admin123（或环境变量中设置的密码）

### Prometheus指标
- 访问：http://localhost:9090
- 查看系统指标和监控数据

### API接口
- 系统信息：http://localhost:8080/api/v1/system/info
- 健康检查：http://localhost:8080/api/v1/system/health

## 🔐 安全建议

1. **修改默认密码**
2. **限制端口访问**
3. **使用HTTPS**
4. **定期备份数据**
5. **监控系统日志**

## 📚 更多信息

- [完整文档](README.md)
- [配置指南](configs/config.yaml)
- [API文档](build-docker.md)
