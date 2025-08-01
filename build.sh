#!/bin/bash

# FRP 配置管理器多架构构建脚本

set -e

PROGRAM_NAME="webfrpc"
VERSION=$(date +"%Y%m%d_%H%M%S")

echo "========================================="
echo "FRP 配置管理器多架构构建"
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
echo "构建版本: ${VERSION}"
echo ""

# 清理旧的构建文件
echo "清理旧的构建文件..."
rm -f ${PROGRAM_NAME}* ${PROGRAM_NAME}-*.tar.gz

# 定义构建目标
declare -A TARGETS=(
    ["linux/amd64"]="x86_64"
    ["linux/arm64"]="arm64"
)

# 构建多架构版本
for target in "${!TARGETS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$target"
    ARCH_NAME="${TARGETS[$target]}"
    OUTPUT_NAME="${PROGRAM_NAME}-${ARCH_NAME}"

    echo "正在构建 ${ARCH_NAME} 版本 (${GOOS}/${GOARCH})..."

    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-s -w -X main.version=${VERSION}" \
        -o ${OUTPUT_NAME} main.go

    if [ $? -ne 0 ]; then
        echo "错误：${ARCH_NAME} 版本构建失败"
        exit 1
    fi

    echo "✅ ${ARCH_NAME} 版本构建成功"
done

echo ""
echo "========================================="
echo "构建完成！"
echo "========================================="

# 显示构建结果
for target in "${!TARGETS[@]}"; do
    ARCH_NAME="${TARGETS[$target]}"
    OUTPUT_NAME="${PROGRAM_NAME}-${ARCH_NAME}"

    if [ -f "${OUTPUT_NAME}" ]; then
        SIZE=$(du -h ${OUTPUT_NAME} | cut -f1)
        echo "${ARCH_NAME} 版本: ${OUTPUT_NAME} (${SIZE})"
    fi
done

echo ""
echo "版本: ${VERSION}"
echo ""
echo "特性："
echo "✅ 包含所有 web 文件"
echo "✅ 可以在任何目录运行"
echo "✅ 单文件部署"
echo "✅ 无需外部依赖"
echo "✅ 支持多架构部署"
echo ""
echo "使用方法："
echo "  x86_64: ./${PROGRAM_NAME}-x86_64"
echo "  arm64:  ./${PROGRAM_NAME}-arm64"
echo ""
echo "访问地址："
echo "  http://localhost:8888"
echo ""

# 测试 x86_64 版本（如果在 x86_64 系统上）
if [ "$(uname -m)" = "x86_64" ] && [ -f "${PROGRAM_NAME}-x86_64" ]; then
    echo "正在测试 x86_64 版本..."
    timeout 3s ./${PROGRAM_NAME}-x86_64 > /dev/null 2>&1 &
    TEST_PID=$!
    sleep 2

    if kill -0 $TEST_PID 2>/dev/null; then
        echo "✅ x86_64 版本测试通过"
        kill $TEST_PID 2>/dev/null || true
    else
        echo "❌ x86_64 版本测试失败"
    fi
fi

# 创建发布包
echo ""
echo "创建发布包..."
PACKAGE_NAME="${PROGRAM_NAME}-${VERSION}"
mkdir -p ${PACKAGE_NAME}

# 复制文件到发布包
for target in "${!TARGETS[@]}"; do
    ARCH_NAME="${TARGETS[$target]}"
    OUTPUT_NAME="${PROGRAM_NAME}-${ARCH_NAME}"

    if [ -f "${OUTPUT_NAME}" ]; then
        cp ${OUTPUT_NAME} ${PACKAGE_NAME}/
        chmod +x ${PACKAGE_NAME}/${OUTPUT_NAME}
    fi
done

# 复制其他文件
cp -r web ${PACKAGE_NAME}/
cp README.md ${PACKAGE_NAME}/
cp start.sh ${PACKAGE_NAME}/
cp install.sh ${PACKAGE_NAME}/

# 创建启动脚本
cat > ${PACKAGE_NAME}/start-auto.sh << 'EOF'
#!/bin/bash
echo "启动 FRP 配置管理器（自动选择架构）..."

# 检测系统架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        PROGRAM="./webfrpc-x86_64"
        ;;
    aarch64|arm64)
        PROGRAM="./webfrpc-arm64"
        ;;
    *)
        echo "错误：不支持的架构 $ARCH"
        exit 1
        ;;
esac

if [ ! -f "$PROGRAM" ]; then
    echo "错误：找不到适用于 $ARCH 架构的程序文件"
    exit 1
fi

echo "检测到架构: $ARCH"
echo "使用程序: $PROGRAM"
echo "请在浏览器中访问: http://localhost:8888"
echo "或者访问: http://$(hostname -I | awk '{print $1}'):8888"
echo ""

$PROGRAM
EOF

chmod +x ${PACKAGE_NAME}/start-auto.sh

# 打包
tar -czf ${PACKAGE_NAME}.tar.gz ${PACKAGE_NAME}
rm -rf ${PACKAGE_NAME}

echo "发布包: ${PACKAGE_NAME}.tar.gz"
echo ""
echo "部署方法："
echo "1. 解压: tar -xzf ${PACKAGE_NAME}.tar.gz"
echo "2. 进入目录: cd ${PACKAGE_NAME}"
echo "3. 自动启动: ./start-auto.sh"
echo "4. 或手动选择: ./webfrpc-x86_64 或 ./webfrpc-arm64"
echo ""
echo "构建完成！程序已准备好部署。"
