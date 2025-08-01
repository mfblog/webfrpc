#!/bin/bash

# FRP 配置管理器嵌入版本构建脚本

set -e

PROGRAM_NAME="webfrpc"
VERSION=$(date +"%Y%m%d_%H%M%S")

echo "========================================="
echo "FRP 配置管理器嵌入版本构建"
echo "========================================="

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "错误：未找到 Go 编译器"
    exit 1
fi

# 检查 web 目录是否存在
if [ ! -d "./web" ]; then
    echo "错误：web 目录不存在！"
    exit 1
fi

echo "Go 版本: $(go version)"
echo "构建时间: $(date)"
echo ""

# 构建嵌入版本
echo "正在构建嵌入版本..."
go build -ldflags "-s -w -X main.version=${VERSION}" -o ${PROGRAM_NAME} main.go

if [ $? -ne 0 ]; then
    echo "错误：构建失败"
    exit 1
fi

# 显示结果
SIZE=$(du -h ${PROGRAM_NAME} | cut -f1)
echo ""
echo "========================================="
echo "构建成功！"
echo "========================================="
echo "程序文件: ${PROGRAM_NAME} (${SIZE})"
echo "版本: ${VERSION}"
echo ""
echo "特性："
echo "✅ 包含所有 web 文件"
echo "✅ 可以在任何目录运行"
echo "✅ 单文件部署"
echo "✅ 无需外部依赖"
echo ""
echo "使用方法："
echo "  ./${PROGRAM_NAME}"
echo ""
echo "访问地址："
echo "  http://localhost:8080"
echo ""

# 测试程序是否能正常启动
echo "正在测试程序..."
timeout 3s ./${PROGRAM_NAME} > /dev/null 2>&1 &
TEST_PID=$!
sleep 2

if kill -0 $TEST_PID 2>/dev/null; then
    echo "✅ 程序测试通过"
    kill $TEST_PID 2>/dev/null || true
else
    echo "❌ 程序测试失败"
fi

echo ""
echo "构建完成！程序已准备好部署。"
