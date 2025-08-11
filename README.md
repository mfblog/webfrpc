# FRP 客户端配置管理器

这是一个使用 Go + Vue3 + Tailwind CSS 开发的 FRP 客户端配置管理程序，提供了友好的 Web 界面来管理 frpc.toml 配置文件。

## 功能特性

- 🌐 **Web 界面管理**：通过浏览器直观地管理 FRP 配置
- 🔧 **完整配置支持**：支持服务器配置、TLS 设置、日志配置等
- 📋 **代理管理**：可以添加、删除、修改代理配置
- 🔄 **自动重启**：配置保存后自动重启 frpc 服务
- ❌ **错误显示**：实时显示配置错误和服务状态
- 💾 **配置备份**：每次保存前自动备份原配置文件
- 🎯 **智能界面**：根据配置文件是否存在自动切换界面
- ⚙️ **设置管理**：在代理界面可通过设置按钮修改基础配置
- 🔒 **TLS 开关**：TLS 配置支持开关控制，关闭时不会生成 TLS 配置
- 🏷️ **智能代理名称**：代理列表显示实际的代理名称而不是序号
- 🔗 **连接状态检查**：保存配置后自动检查服务器连接状态
- 📋 **实时日志查看**：支持查看 systemctl 服务日志，自动刷新
- 📁 **代理卡片折叠**：代理配置支持折叠/展开，提升界面整洁度
- ⚙️ **systemctl 集成**：使用 systemctl 管理 frpc 服务，更加稳定可靠
- 🚀 **自动安装检查**：启动时自动检查系统状态，缺少组件时引导安装
- 📦 **一键安装更新**：支持自动下载安装最新版本的 frpc 客户端
- 🔍 **版本检查**：自动检查并提示 frpc 客户端更新
- 📁 **文件嵌入打包**：支持将 web 文件嵌入到程序中，实现单文件部署

## 快速开始

### 最简单的用法

```bash
mkdir -p /usr/local/frp
```

### 将 install-service.sh 和 对应架构的程序放到 /usr/local/frp

### 给权限

```bash
chmod 755 install-service.sh webfrpc-x86_64
```

### 运行

```bash
./install-service.sh
```

### 浏览器打开 http://ip:9696 web 管理界面开始使用就可以了

```bash
可以git clone https://github.com/mfblog/webfrpc.git
进入目录。使用 build.sh 自己编译
```

**首次运行**：程序会自动进行系统检查：

- 检查并创建 `/usr/local/frp` 目录
- 检查 frpc 客户端是否存在
- 检查 systemd 服务是否配置
- 如果缺少组件，会在 Web 界面提供安装选项

### 3. 访问 Web 界面

程序启动后，在浏览器中访问：

```
# 本地访问
http://localhost:9696

# 网络访问（替换为实际 IP）
http://192.168.1.100:9696
```

### 4. 配置管理

#### 首次使用

**如果 frpc 客户端不存在**：

- 程序会显示安装界面
- 显示系统状态检查结果
- 点击"安装 FRP 客户端"自动下载安装最新版本

**如果配置文件不存在**：

- 程序会自动显示初始配置界面
- 需要完成基础配置、TLS 配置、日志配置
- 点击"创建配置并启动服务"完成初始化

#### 日常使用（配置文件已存在）

- 程序会自动显示代理配置界面
- 代理列表显示实际的代理名称（如 "WSL"、"SSH"）
- 代理配置卡片默认折叠，点击可展开编辑
- 新添加的代理默认展开，保存后自动折叠
- 可以直接添加、删除、修改代理规则
- 点击右上角"日志"按钮查看 systemctl 服务日志（自动实时刷新）
- 点击右上角"设置"按钮可修改基础配置、TLS 配置、日志配置
- TLS 配置支持开关控制，关闭时不会在配置文件中生成 TLS 部分
- 保存配置后会自动检查服务器连接状态并显示结果
- 成功和连接状态提示会在 5 秒后自动消失
- 如果有新版本可用，会在界面顶部显示更新提示

### 5. 保存配置

点击"保存配置并重启服务"按钮，程序会：

1. 验证配置格式
2. 备份原配置文件
3. 保存新配置到 frpc.toml
4. 自动重启 frpc 服务
5. 显示操作结果

## 目录结构

```
/usr/local/frp/
├── frpc                    # frpc 可执行文件
├── frpc.toml              # frpc 配置文件
├── ssl/                   # SSL 证书目录
│   ├── ca.crt
│   ├── client.crt
│   └── client.key
├── main.go                # Go 后端源码
├── web/
│   └── index.html         # Vue3 前端页面
├── frpc-config-manager    # 编译后的程序
├── start.sh              # 启动脚本
└── README.md             # 说明文档
```

## API 接口

### GET /api/config

获取当前配置

### POST /api/config

保存配置并重启服务

### GET /api/status

获取服务状态

### GET /api/check

检查配置文件是否存在

### GET /api/logs

获取 systemctl 服务日志

### GET /api/service-status

获取 systemctl 服务状态

### GET /api/system-status

获取系统完整状态（目录、frpc、服务、配置、版本信息）

### POST /api/install-frpc

安装 frpc 客户端

### POST /api/update-frpc

更新 frpc 客户端到最新版本

## 构建和打包

### 构建嵌入版本（推荐）

```bash
./build-embedded.sh
```

- 将 web 文件嵌入到程序中
- 生成单个可执行文件
- 可在任何目录运行
- 无需外部依赖

### 构建普通版本

```bash
go build -o webfrpc main.go
```

- 需要 web 目录在同一位置
- 适合开发调试

### 手动构建嵌入版本

```bash
go build -ldflags "-s -w" -o webfrpc main.go
```

- `-s -w` 参数用于减小程序体积
- 自动嵌入 web/\* 目录下的所有文件

## 技术栈

- **后端**：Go + Gin + go-toml
- **前端**：Vue3 + Tailwind CSS + Axios
- **配置格式**：TOML

## 注意事项

1. 确保系统已安装并配置 frpc.service systemd 服务
2. 确保有足够权限执行 systemctl 命令（可能需要 sudo 权限）
3. 确保有足够权限读写配置文件
4. 配置保存前会自动备份，备份文件格式为 `frpc.toml.backup.YYYYMMDDHHMMSS`
5. 服务重启使用 `systemctl restart frpc.service`
6. 日志查看使用 `journalctl -u frpc.service`

## 故障排除

### 程序无法启动

- 检查 Go 环境是否正确安装
- 检查端口 8888 是否被占用

### frpc 服务无法启动

- 检查 frpc 可执行文件是否存在
- 检查配置文件格式是否正确
- 检查证书文件路径是否正确

### 无法连接到服务器

- 检查服务器地址和端口是否正确
- 检查认证令牌是否正确
- 检查网络连接是否正常

## 许可证

本项目版权归作者所有，仅限个人学习与研究使用，禁止用于任何商业用途。
For personal learning and research only. Commercial use is strictly prohibited.
