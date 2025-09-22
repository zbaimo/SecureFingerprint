# 🛡️ 防火墙控制器 (Firewall Controller)

一个基于用户指纹的智能访问控制系统，提供实时行为分析、风险评估和自动防护功能。

## ✨ 核心特性

### 🎯 智能识别
- **用户指纹生成**: 基于IP、User-Agent、HTTP头等信息生成稳定的用户标识
- **设备类型检测**: 自动识别移动设备、桌面浏览器、机器人等
- **网络类型分析**: 区分内网、代理、移动网络等访问来源

### 📊 行为分析
- **访问模式识别**: 检测异常访问频率、路径模式、时间分布
- **机器人检测**: 多维度识别爬虫、扫描器等自动化工具
- **风险评估**: 实时计算用户风险等级，支持自定义评分规则

### 🔒 访问控制
- **多层防护**: 支持限速、人机验证、临时封禁等多种处理方式
- **动态调整**: 根据用户行为动态调整访问权限
- **白名单机制**: 支持可信用户白名单，避免误杀

### 📈 可视化管理
- **实时监控**: 直观的仪表板显示系统运行状态
- **日志分析**: 详细的访问日志和统计分析
- **配置管理**: 灵活的规则配置和参数调优

## 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   Mobile App    │    │   API Client    │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼───────────────┐
                    │         Nginx               │
                    │    (Reverse Proxy)          │
                    └─────────────┬───────────────┘
                                 │
                    ┌─────────────▼───────────────┐
                    │    Firewall Controller      │
                    │                             │
                    │  ┌─────────┐ ┌─────────┐   │
                    │  │Collector│ │Analyzer │   │
                    │  └─────────┘ └─────────┘   │
                    │  ┌─────────┐ ┌─────────┐   │
                    │  │ Scorer  │ │ Limiter │   │
                    │  └─────────┘ └─────────┘   │
                    └─────────────┬───────────────┘
                                 │
                    ┌─────────────▼───────────────┐
                    │        Storage              │
                    │                             │
                    │  ┌─────────┐ ┌─────────┐   │
                    │  │  Redis  │ │  MySQL  │   │
                    │  └─────────┘ └─────────┘   │
                    └─────────────────────────────┘
```

## 🚀 快速开始

### 环境要求

- Docker & Docker Compose
- Go 1.21+ (开发环境)
- Node.js 18+ (前端开发)

### 一键部署

```bash
# 克隆项目
git clone https://github.com/your-org/firewall-controller.git
cd firewall-controller

# 启动完整环境
make up

# 访问系统
open http://localhost
```

### 开发环境

```bash
# 安装开发工具
make install-tools

# 启动基础服务
make dev-up

# 开发模式运行
make dev
```

## 📖 使用指南

### 系统配置

主要配置文件位于 `configs/config.yaml`：

```yaml
security:
  scoring:
    initial_score: 100        # 初始分数
    ban_threshold: 0          # 封禁阈值
    bot_penalty: -15          # 机器人扣分
  
  limiter:
    max_requests_per_window: 100  # 窗口最大请求数
    ban_duration: 3600s          # 封禁时长
  
  analyzer:
    suspicious_request_threshold: 50  # 可疑请求阈值
    bot_detection_enabled: true       # 启用机器人检测
```

### API接口

系统提供完整的REST API：

- **系统信息**: `GET /api/v1/system/info`
- **访问日志**: `GET /api/v1/logs`
- **用户分数**: `GET /api/v1/score/{fingerprint}`
- **风控规则**: `GET /api/v1/rule/ban`

详细API文档请查看 [API文档](docs/api.md)

### 集成指南

#### 中间件集成

```go
import "firewall-controller/pkg/middleware"

// Gin框架
router.Use(middleware.FirewallMiddleware())

// 原生HTTP
http.Handle("/", middleware.FirewallHandler(yourHandler))
```

#### 自定义规则

```go
// 自定义评分规则
scorer.AddRule("custom", func(info *AccessInfo) int {
    if strings.Contains(info.Path, "/admin") {
        return -10
    }
    return 0
})
```

## 🔧 配置说明

### 评分系统

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `initial_score` | 100 | 新用户初始分数 |
| `normal_access_bonus` | +1 | 正常访问加分 |
| `bot_penalty` | -15 | 机器人行为扣分 |
| `frequent_request_penalty` | -10 | 频繁请求扣分 |

### 限制器配置

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `rate_limit_window` | 60s | 限速时间窗口 |
| `max_requests_per_window` | 100 | 窗口最大请求数 |
| `ban_duration` | 3600s | 封禁持续时间 |

### 行为分析

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `suspicious_request_threshold` | 50 | 可疑请求阈值 |
| `path_repeat_threshold` | 10 | 路径重复访问阈值 |
| `bot_detection_enabled` | true | 启用机器人检测 |

## 📊 监控指标

系统提供丰富的监控指标：

- **请求统计**: 总请求数、成功率、错误率
- **用户分析**: 活跃用户数、新用户数、风险用户数
- **安全事件**: 封禁数、拦截数、误报率
- **性能指标**: 响应时间、吞吐量、资源使用率

## 🛠️ 开发指南

### 项目结构

```
firewall-controller/
├── cmd/server/          # 程序入口
├── internal/            # 内部核心逻辑
│   ├── collector/       # 数据采集
│   ├── fingerprint/     # 指纹生成
│   ├── scorer/          # 打分系统
│   ├── analyzer/        # 行为分析
│   ├── limiter/         # 访问限制
│   └── storage/         # 数据存储
├── api/                 # WebUI接口
├── webui/               # Vue3前端
├── pkg/                 # 公共库
├── configs/             # 配置文件
├── docker/              # Docker配置
└── docs/                # 文档
```

### 开发流程

```bash
# 1. 创建功能分支
git checkout -b feature/new-feature

# 2. 开发和测试
make dev
make test

# 3. 代码检查
make lint
make security

# 4. 提交代码
git commit -m "feat: add new feature"
git push origin feature/new-feature
```

### 测试

```bash
# 运行所有测试
make test

# 查看覆盖率
make coverage

# 性能测试
make bench
```

## 🐳 Docker部署

### 单机部署

```bash
# 构建镜像
make docker-build

# 启动服务
docker-compose up -d
```

### 集群部署

支持Kubernetes部署，配置文件位于 `k8s/` 目录：

```bash
kubectl apply -f k8s/
```

## 📈 性能优化

### 系统调优

1. **Redis配置**
   - 合理设置内存限制
   - 启用持久化
   - 配置集群模式

2. **MySQL优化**
   - 索引优化
   - 查询缓存
   - 读写分离

3. **应用层优化**
   - 连接池配置
   - 缓存策略
   - 异步处理

### 扩容方案

- **水平扩展**: 支持多实例部署
- **缓存分层**: Redis + 本地缓存
- **数据库分片**: 按时间或用户分片

## 🔐 安全考虑

### 数据保护

- 用户指纹采用哈希加盐
- 敏感数据加密存储
- 访问日志定期清理

### 访问控制

- API接口鉴权
- 管理后台权限控制
- 操作审计日志

### 防护机制

- 防止配置篡改
- 限制管理接口访问
- 异常行为告警

## 📚 文档

- [安装指南](docs/installation.md)
- [配置参考](docs/configuration.md)
- [API文档](docs/api.md)
- [开发指南](docs/development.md)
- [部署指南](docs/deployment.md)
- [故障排查](docs/troubleshooting.md)

## 🤝 贡献

欢迎提交Issue和Pull Request！

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 📄 许可证

本项目采用 [MIT许可证](LICENSE)

## 🙏 致谢

感谢所有贡献者和开源项目的支持！

---

<div align="center">
  <p>如果这个项目对你有帮助，请给个 ⭐️ 支持一下！</p>
</div>
