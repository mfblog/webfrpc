#!/bin/bash

# FRP 配置管理器安装脚本

set -e

INSTALL_DIR="/usr/local/frp"
PROGRAM_NAME="webfrpc"

echo "========================================="
echo "FRP 配置管理器安装脚本"
echo "========================================="

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then
    echo "错误：请使用 root 用户运行此脚本"
    echo "使用命令: sudo $0"
    exit 1
fi

# 检查系统依赖
echo "检查系统依赖..."

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "错误：未找到 Go 编译器"
    echo "请先安装 Go: https://golang.org/dl/"
    exit 1
fi

# 检查 curl 是否安装
if ! command -v curl &> /dev/null; then
    echo "正在安装 curl..."
    if command -v apt-get &> /dev/null; then
        apt-get update && apt-get install -y curl
    elif command -v yum &> /dev/null; then
        yum install -y curl
    else
        echo "错误：无法自动安装 curl，请手动安装"
        exit 1
    fi
fi

# 检查 systemctl 是否可用
if ! command -v systemctl &> /dev/null; then
    echo "警告：systemctl 不可用，某些功能可能无法正常工作"
fi

# 创建安装目录
echo "创建安装目录: $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

# 复制文件
echo "复制程序文件..."
if [ -f "./main.go" ]; then
    cp ./main.go "$INSTALL_DIR/"
fi

if [ -f "./go.mod" ]; then
    cp ./go.mod "$INSTALL_DIR/"
fi

if [ -f "./go.sum" ]; then
    cp ./go.sum "$INSTALL_DIR/"
fi

if [ -d "./web" ]; then
    cp -r ./web "$INSTALL_DIR/"
else
    echo "错误：找不到 web 目录"
    exit 1
fi

if [ -f "./start.sh" ]; then
    cp ./start.sh "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/start.sh"
fi

if [ -f "./README.md" ]; then
    cp ./README.md "$INSTALL_DIR/"
fi

# 进入安装目录
cd "$INSTALL_DIR"

# 构建程序
echo "构建程序..."
go build -o "$PROGRAM_NAME" main.go

if [ $? -ne 0 ]; then
    echo "错误：程序构建失败"
    exit 1
fi

# 设置权限
chmod +x "$PROGRAM_NAME"
chmod +x start.sh

echo ""
echo "========================================="
echo "安装完成！"
echo "========================================="
echo "安装目录: $INSTALL_DIR"
echo "程序文件: $INSTALL_DIR/$PROGRAM_NAME"
echo ""
echo "启动方法："
echo "1. 使用启动脚本（推荐）:"
echo "   cd $INSTALL_DIR && ./start.sh"
echo ""
echo "2. 直接运行程序:"
echo "   cd $INSTALL_DIR && ./$PROGRAM_NAME"
echo ""
echo "3. 创建系统服务（可选）:"
echo "   systemctl enable --now frp-config-manager"
echo ""
echo "访问地址："
echo "- 本地访问: http://localhost:8888"
echo "- 网络访问: http://$(hostname -I | awk '{print $1}'):8888"
echo ""
echo "注意事项："
echo "- 程序会自动检查并安装 frpc 客户端"
echo "- 首次运行时会创建必要的系统服务"
echo "- 请确保防火墙允许 8888 端口访问"
echo ""
