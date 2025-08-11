# FRP 配置管理器部署指南

## 快速部署（推荐）

### 方法一：使用预编译版本 + install-service.sh

1. **准备文件**
   ```bash
   # 创建工作目录
   mkdir -p /tmp/webfrpc-deploy
   cd /tmp/webfrpc-deploy
   
   # 编译多架构版本
   GOOS=linux GOARCH=amd64 go build -o webfrpc-x86_64 main.go
   GOOS=linux GOARCH=arm64 go build -o webfrpc-arm64 main.go
   ```

2. **部署到目标目录**
   ```bash
   # 创建安装目录
   sudo mkdir -p /usr/local/frp
   
   # 复制程序文件
   sudo cp webfrpc-x86_64 /usr/local/frp/
   sudo cp webfrpc-arm64 /usr/local/frp/
   sudo cp install-service.sh /usr/local/frp/
   
   # 设置执行权限
   sudo chmod +x /usr/local/frp/webfrpc-*
   sudo chmod +x /usr/local/frp/install-service.sh
   ```

3. **安装系统服务**
   ```bash
   cd /usr/local/frp
   sudo ./install-service.sh
   ```

4. **访问 Web 界面**
   ```
   http://localhost:9696
   ```

### 方法二：使用 build.sh 自动构建

1. **运行构建脚本**
   ```bash
   ./build.sh
   ```

2. **解压发布包**
   ```bash
   tar -xzf webfrpc-*.tar.gz
   cd webfrpc-*
   ```

3. **安装系统服务**
   ```bash
   sudo cp webfrpc-* /usr/local/frp/
   sudo cp install-service.sh /usr/local/frp/
   sudo chmod +x /usr/local/frp/webfrpc-*
   sudo chmod +x /usr/local/frp/install-service.sh
   cd /usr/local/frp
   sudo ./install-service.sh
   ```

## 架构支持

install-service.sh 脚本会自动检测系统架构：

- **x86_64 系统**: 使用 `webfrpc-x86_64`
- **aarch64/arm64 系统**: 使用 `webfrpc-arm64`

## 服务管理

安装完成后，可以使用以下命令管理服务：

```bash
# 查看服务状态
sudo systemctl status webfrpc

# 查看服务日志
sudo journalctl -u webfrpc -f

# 重启服务
sudo systemctl restart webfrpc

# 停止服务
sudo systemctl stop webfrpc

# 启动服务
sudo systemctl start webfrpc

# 禁用开机自启动
sudo systemctl disable webfrpc

# 启用开机自启动
sudo systemctl enable webfrpc
```

## 故障排除

### 1. 程序文件不存在
```
错误：找不到程序文件 /usr/local/frp/webfrpc-arm64
```

**解决方案**: 确保已经编译并复制了对应架构的程序文件。

### 2. 权限问题
```
错误：请使用 root 用户运行此脚本
```

**解决方案**: 使用 `sudo` 运行脚本。

### 3. 服务启动失败

**检查日志**:
```bash
sudo journalctl -u webfrpc --no-pager
```

**常见原因**:
- 端口 9696 被占用
- 程序文件权限不正确
- 依赖文件缺失

## 卸载

如需卸载服务：

```bash
# 停止并禁用服务
sudo systemctl stop webfrpc
sudo systemctl disable webfrpc

# 删除服务文件
sudo rm /etc/systemd/system/webfrpc.service

# 重新加载 systemd
sudo systemctl daemon-reload

# 删除程序文件（可选）
sudo rm -rf /usr/local/frp
```

## 更新程序

1. **停止服务**
   ```bash
   sudo systemctl stop webfrpc
   ```

2. **替换程序文件**
   ```bash
   sudo cp new-webfrpc-x86_64 /usr/local/frp/webfrpc-x86_64
   sudo cp new-webfrpc-arm64 /usr/local/frp/webfrpc-arm64
   sudo chmod +x /usr/local/frp/webfrpc-*
   ```

3. **启动服务**
   ```bash
   sudo systemctl start webfrpc
   ```
