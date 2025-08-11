#!/bin/bash

# FRP 配置管理器系统服务安装脚本

set -e

SERVICE_NAME="webfrpc"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
INSTALL_DIR="/usr/local/frp"

echo "========================================="
echo "FRP 配置管理器系统服务安装"
echo "========================================="

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then
    echo "错误：请使用 root 用户运行此脚本"
    echo "使用命令: sudo $0"
    exit 1
fi

# 检测系统架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        PROGRAM_PATH="${INSTALL_DIR}/webfrpc-x86_64"
        ;;
    aarch64|arm64)
        PROGRAM_PATH="${INSTALL_DIR}/webfrpc-arm64"
        ;;
    *)
        echo "错误：不支持的架构 $ARCH"
        echo "支持的架构: x86_64, aarch64, arm64"
        exit 1
        ;;
esac

echo "检测到架构: $ARCH"
echo "程序路径: $PROGRAM_PATH"

# 检查程序文件是否存在
if [ ! -f "$PROGRAM_PATH" ]; then
    echo "错误：找不到程序文件 $PROGRAM_PATH"
    echo "请确保已经编译并放置了对应架构的程序文件"
    exit 1
   fi
   
   # 创建或更新符号链接，指向特定架构的程序
   SYMLINK_PATH="${INSTALL_DIR}/webfrpc"
   echo "创建符号链接: ${SYMLINK_PATH} -> ${PROGRAM_PATH}"
   ln -sf "$PROGRAM_PATH" "$SYMLINK_PATH"
   
   # 检查 systemctl 是否可用
   if ! command -v systemctl &> /dev/null; then
    echo "错误：systemctl 不可用，此系统可能不支持 systemd"
    exit 1
fi

# 停止现有服务（如果存在）
if systemctl is-active --quiet $SERVICE_NAME; then
    echo "停止现有服务..."
    systemctl stop $SERVICE_NAME
fi

# 创建服务文件
echo "创建系统服务文件..."
cat > $SERVICE_FILE << EOF
[Unit]
Description=FRP Configuration web
Documentation=https://github.com/mfblog/webfrpc
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Restart=on-failure
RestartSec=5s
ExecStart=$SYMLINK_PATH
WorkingDirectory=$INSTALL_DIR
Environment=GIN_MODE=release
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF

echo "创建 frpc 系统服务文件..."
FRPC_SERVICE_FILE="/etc/systemd/system/frpc.service"
FRPC_PATH="${INSTALL_DIR}/frpc"
FRPC_CONFIG_PATH="${INSTALL_DIR}/frpc.toml"

cat > $FRPC_SERVICE_FILE << EOF
[Unit]
Description=FRP Client
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Restart=on-failure
RestartSec=5s
ExecStart=$FRPC_PATH -c $FRPC_CONFIG_PATH
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF

# 重新加载 systemd 配置
echo "重新加载 systemd 配置..."
systemctl daemon-reload

# 启用 webfrpc 服务
echo "启用 ${SERVICE_NAME} 服务（开机自启动）..."
systemctl enable $SERVICE_NAME

# 启用 frpc 服务
echo "启用 frpc.service 服务（开机自启动）..."
systemctl enable frpc.service

# 启动服务
echo "启动服务..."
systemctl start $SERVICE_NAME

# 等待服务启动
sleep 3

# 检查服务状态
if systemctl is-active --quiet $SERVICE_NAME; then
    echo ""
    echo "========================================="
    echo "安装成功！"
    echo "========================================="
    echo "服务名称: $SERVICE_NAME"
    echo "服务状态: $(systemctl is-active $SERVICE_NAME)"
    echo "访问地址: http://localhost:9696"
    echo "网络访问: http://$(hostname -I | awk '{print $1}'):9696"
    echo ""
    echo "管理命令："
    echo "  查看状态: sudo systemctl status $SERVICE_NAME"
    echo "  查看日志: sudo journalctl -u $SERVICE_NAME -f"
    echo "  重启服务: sudo systemctl restart $SERVICE_NAME"
    echo "  停止服务: sudo systemctl stop $SERVICE_NAME"
    echo "  禁用服务: sudo systemctl disable $SERVICE_NAME"
    echo ""
    echo "服务已启动并设置为开机自启动！"
else
    echo ""
    echo "========================================="
    echo "安装失败！"
    echo "========================================="
    echo "服务启动失败，请检查日志："
    echo "sudo journalctl -u $SERVICE_NAME --no-pager"
    exit 1
fi
