@echo off
REM Windows批处理脚本 - 构建和推送Docker镜像

setlocal enabledelayedexpansion

REM 配置变量
set DOCKER_USERNAME=%DOCKER_USERNAME%
if "%DOCKER_USERNAME%"=="" set DOCKER_USERNAME=zbaimo
set IMAGE_NAME=firewall-controller
set VERSION=%VERSION%
if "%VERSION%"=="" set VERSION=v1.0.0
set PLATFORM=linux/amd64,linux/arm64

echo.
echo 🐳 开始构建防火墙控制器Docker镜像...
echo.

REM 检查Docker是否安装
docker --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker 未安装或未启动
    exit /b 1
)

REM 检查Docker Buildx
docker buildx version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker Buildx 未安装
    exit /b 1
)

echo ✅ Docker 环境检查完成

REM 登录Docker Hub
echo.
echo 📝 登录Docker Hub...
if not "%DOCKER_PASSWORD%"=="" (
    echo %DOCKER_PASSWORD% | docker login -u %DOCKER_USERNAME% --password-stdin
) else (
    docker login -u %DOCKER_USERNAME%
)

if errorlevel 1 (
    echo ❌ Docker Hub 登录失败
    exit /b 1
)

echo ✅ Docker Hub 登录成功

REM 设置Buildx
echo.
echo 🔧 设置Docker Buildx...
docker buildx create --name multiarch --driver docker-container --use 2>nul || echo Buildx实例已存在
docker buildx inspect --bootstrap

REM 获取构建信息
for /f "tokens=*" %%i in ('powershell -command "Get-Date -UFormat '%%Y-%%m-%%dT%%H:%%M:%%SZ'"') do set BUILD_TIME=%%i

REM 尝试获取Git提交信息
git rev-parse --short HEAD >nul 2>&1
if errorlevel 1 (
    set GIT_COMMIT=unknown
) else (
    for /f "tokens=*" %%i in ('git rev-parse --short HEAD') do set GIT_COMMIT=%%i
)

REM 构建信息
set FULL_IMAGE_NAME=%DOCKER_USERNAME%/%IMAGE_NAME%

echo.
echo 📦 构建信息:
echo   镜像名称: %FULL_IMAGE_NAME%
echo   版本标签: %VERSION%, latest
echo   构建平台: %PLATFORM%
echo   构建时间: %BUILD_TIME%
echo   Git提交: %GIT_COMMIT%
echo.

REM 构建并推送镜像
echo 🏗️ 开始构建多平台镜像...
docker buildx build ^
    --platform %PLATFORM% ^
    --build-arg VERSION=%VERSION% ^
    --build-arg BUILD_TIME=%BUILD_TIME% ^
    --build-arg GIT_COMMIT=%GIT_COMMIT% ^
    --tag %FULL_IMAGE_NAME%:%VERSION% ^
    --tag %FULL_IMAGE_NAME%:latest ^
    --push ^
    .

if errorlevel 1 (
    echo ❌ 镜像构建失败
    goto cleanup
)

echo ✅ 镜像构建和推送完成

REM 验证镜像
echo.
echo 🔍 验证镜像...
docker pull %FULL_IMAGE_NAME%:latest

if errorlevel 1 (
    echo ❌ 镜像验证失败
    goto cleanup
)

echo ✅ 镜像验证完成

REM 显示使用说明
echo.
echo ========================================
echo 🛡️  防火墙控制器镜像构建完成
echo ========================================
echo.
echo 📦 镜像信息:
echo   - 名称: %FULL_IMAGE_NAME%
echo   - 版本: %VERSION%, latest
echo   - 平台: %PLATFORM%
echo.
echo 🚀 使用方法:
echo.
echo 1. 快速启动:
echo    docker run -d -p 8080:8080 %FULL_IMAGE_NAME%:latest
echo.
echo 2. 使用Docker Compose:
echo    curl -O https://raw.githubusercontent.com/your-org/firewall-controller/main/docker-compose.yml
echo    docker-compose up -d
echo.
echo 3. 环境变量配置:
echo    docker run -d \
echo      -p 8080:8080 \
echo      -e REDIS_HOST=redis \
echo      -e MYSQL_HOST=mysql \
echo      %FULL_IMAGE_NAME%:latest
echo.
echo 4. 数据持久化:
echo    docker run -d \
echo      -p 8080:8080 \
echo      -v /host/logs:/app/logs \
echo      -v /host/data:/app/data \
echo      %FULL_IMAGE_NAME%:latest
echo.
echo 📚 更多信息:
echo   - 文档: https://github.com/your-org/firewall-controller
echo   - Docker Hub: https://hub.docker.com/r/%FULL_IMAGE_NAME%
echo.
echo ========================================

goto end

:cleanup
echo.
echo 🧹 清理构建环境...
docker buildx rm multiarch 2>nul || echo 清理完成
exit /b 1

:end
echo.
echo 🧹 清理构建环境...
docker buildx rm multiarch 2>nul || echo 清理完成
echo.
echo 🎉 构建完成！
pause
