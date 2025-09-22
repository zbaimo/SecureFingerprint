# Makefile for Firewall Controller

# 变量定义
APP_NAME = firewall-controller
VERSION ?= v1.0.0
DOCKER_IMAGE = $(APP_NAME):$(VERSION)
DOCKER_REGISTRY ?= localhost:5000

# Go相关变量
GO_VERSION = 1.21
GO_BUILD_FLAGS = -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"
GO_FILES = $(shell find . -name "*.go" -type f)

# 默认目标
.PHONY: all
all: clean build

# 清理构建产物
.PHONY: clean
clean:
	@echo "清理构建产物..."
	@rm -rf build/
	@rm -f $(APP_NAME)
	@docker rmi $(DOCKER_IMAGE) 2>/dev/null || true

# 安装依赖
.PHONY: deps
deps:
	@echo "安装Go依赖..."
	@go mod download
	@go mod tidy

# 代码检查
.PHONY: lint
lint:
	@echo "运行代码检查..."
	@golangci-lint run ./...

# 运行测试
.PHONY: test
test:
	@echo "运行测试..."
	@go test -v -race -coverprofile=coverage.out ./...

# 查看测试覆盖率
.PHONY: coverage
coverage: test
	@echo "生成测试覆盖率报告..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 构建二进制文件
.PHONY: build
build: deps
	@echo "构建应用程序..."
	@mkdir -p build
	@go build $(GO_BUILD_FLAGS) -o build/$(APP_NAME) ./cmd/server

# 构建前端
.PHONY: build-frontend
build-frontend:
	@echo "构建前端..."
	@cd webui && npm ci && npm run build

# 构建所有组件
.PHONY: build-all
build-all: build-frontend build

# 本地运行
.PHONY: run
run: build
	@echo "启动应用程序..."
	@./build/$(APP_NAME)

# 开发模式运行
.PHONY: dev
dev:
	@echo "开发模式启动..."
	@go run ./cmd/server

# 构建Docker镜像
.PHONY: docker-build
docker-build:
	@echo "构建Docker镜像..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker镜像构建完成: $(DOCKER_IMAGE)"

# 推送Docker镜像
.PHONY: docker-push
docker-push: docker-build
	@echo "推送Docker镜像到仓库..."
	@docker tag $(DOCKER_IMAGE) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)

# 启动开发环境
.PHONY: dev-up
dev-up:
	@echo "启动开发环境..."
	@docker-compose -f docker-compose.yml up -d redis mysql
	@echo "等待数据库启动..."
	@sleep 10
	@echo "开发环境已启动"

# 停止开发环境
.PHONY: dev-down
dev-down:
	@echo "停止开发环境..."
	@docker-compose down

# 启动完整环境
.PHONY: up
up:
	@echo "启动完整环境..."
	@docker-compose up -d
	@echo "环境启动完成，访问 http://localhost"

# 停止完整环境
.PHONY: down
down:
	@echo "停止完整环境..."
	@docker-compose down
	@echo "环境已停止"

# 重启环境
.PHONY: restart
restart: down up

# 查看日志
.PHONY: logs
logs:
	@docker-compose logs -f firewall-controller

# 查看所有服务日志
.PHONY: logs-all
logs-all:
	@docker-compose logs -f

# 进入应用容器
.PHONY: shell
shell:
	@docker-compose exec firewall-controller sh

# 数据库备份
.PHONY: db-backup
db-backup:
	@echo "备份数据库..."
	@mkdir -p backups
	@docker-compose exec mysql mysqldump -u root -pfirewall_root_password firewall_controller > backups/firewall_controller_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "数据库备份完成"

# 数据库恢复
.PHONY: db-restore
db-restore:
	@echo "请指定备份文件: make db-restore FILE=backups/xxx.sql"
	@if [ -z "$(FILE)" ]; then echo "错误: 请指定 FILE 参数"; exit 1; fi
	@docker-compose exec -T mysql mysql -u root -pfirewall_root_password firewall_controller < $(FILE)
	@echo "数据库恢复完成"

# 生成API文档
.PHONY: docs
docs:
	@echo "生成API文档..."
	@swag init -g cmd/server/main.go -o docs/
	@echo "API文档已生成到 docs/ 目录"

# 安装开发工具
.PHONY: install-tools
install-tools:
	@echo "安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "开发工具安装完成"

# 格式化代码
.PHONY: fmt
fmt:
	@echo "格式化代码..."
	@go fmt ./...
	@goimports -w $(GO_FILES)

# 更新依赖
.PHONY: update-deps
update-deps:
	@echo "更新依赖..."
	@go get -u ./...
	@go mod tidy

# 安全检查
.PHONY: security
security:
	@echo "运行安全检查..."
	@gosec ./...

# 性能测试
.PHONY: bench
bench:
	@echo "运行性能测试..."
	@go test -bench=. -benchmem ./...

# 构建发布包
.PHONY: release
release: clean build-all
	@echo "构建发布包..."
	@mkdir -p release
	@cp -r build/ release/
	@cp -r configs/ release/
	@cp docker-compose.yml release/
	@cp README.md release/
	@tar -czf release/$(APP_NAME)-$(VERSION).tar.gz -C release .
	@echo "发布包已生成: release/$(APP_NAME)-$(VERSION).tar.gz"

# 健康检查
.PHONY: health
health:
	@echo "检查服务健康状态..."
	@curl -f http://localhost:8080/api/v1/system/health || echo "服务未响应"

# 监控指标
.PHONY: metrics
metrics:
	@echo "获取监控指标..."
	@curl -s http://localhost:8080/api/v1/metrics || echo "指标接口未响应"

# 显示帮助信息
.PHONY: help
help:
	@echo "Firewall Controller Makefile"
	@echo ""
	@echo "可用命令:"
	@echo "  build          构建应用程序"
	@echo "  build-frontend 构建前端"
	@echo "  build-all      构建所有组件"
	@echo "  clean          清理构建产物"
	@echo "  deps           安装依赖"
	@echo "  dev            开发模式运行"
	@echo "  dev-up         启动开发环境"
	@echo "  dev-down       停止开发环境"
	@echo "  docker-build   构建Docker镜像"
	@echo "  docker-push    推送Docker镜像"
	@echo "  up             启动完整环境"
	@echo "  down           停止完整环境"
	@echo "  restart        重启环境"
	@echo "  logs           查看应用日志"
	@echo "  logs-all       查看所有服务日志"
	@echo "  test           运行测试"
	@echo "  coverage       生成测试覆盖率报告"
	@echo "  lint           代码检查"
	@echo "  fmt            格式化代码"
	@echo "  security       安全检查"
	@echo "  bench          性能测试"
	@echo "  release        构建发布包"
	@echo "  health         健康检查"
	@echo "  help           显示此帮助信息"
