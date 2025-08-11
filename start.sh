#!/bin/bash

# FRP 配置管理器启动脚本

echo "正在启动 FRP 配置管理器..."

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "工作目录: $SCRIPT_DIR"

# 检查是否已经构建了程序
PROGRAM_NAME="webfrpc"
if [ ! -f "./$PROGRAM_NAME" ]; then
    echo "正在构建程序..."
    go build -o "$PROGRAM_NAME" main.go
    if [ $? -ne 0 ]; then
        echo "构建失败！"
        exit 1
    fi
    echo "程序构建成功: $PROGRAM_NAME"
fi

# 检查 web 目录是否存在
if [ ! -d "./web" ]; then
    echo "错误：web 目录不存在！"
    echo "请确保 web 目录与程序在同一目录下。"
    exit 1
fi

# 启动程序
echo "启动配置管理器..."
echo "请在浏览器中访问: http://localhost:9696"
echo "或者访问: http://$(hostname -I | awk '{print $1}'):9696"
echo ""
./"$PROGRAM_NAME"
