package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
)

//go:embed web/*
var webFS embed.FS

// Config 结构体定义
type Config struct {
	ServerAddr string    `toml:"serverAddr" json:"serverAddr"`
	ServerPort int       `toml:"serverPort" json:"serverPort"`
	Transport  Transport `toml:"transport" json:"transport"`
	Auth       Auth      `toml:"auth" json:"auth"`
	Log        Log       `toml:"log" json:"log"`
	Proxies    []Proxy   `toml:"proxies" json:"proxies"`
}

type Transport struct {
	Protocol string `toml:"protocol" json:"protocol"`
	TLS      TLS    `toml:"tls" json:"tls"`
}

type TLS struct {
	CertFile      string `toml:"certFile" json:"certFile"`
	KeyFile       string `toml:"keyFile" json:"keyFile"`
	TrustedCaFile string `toml:"trustedCaFile" json:"trustedCaFile"`
}

type Auth struct {
	Token string `toml:"token" json:"token"`
}

type Log struct {
	To      string `toml:"to" json:"to"`
	Level   string `toml:"level" json:"level"`
	MaxDays int    `toml:"maxDays" json:"maxDays"`
}

type Proxy struct {
	Name       string `toml:"name" json:"name"`
	Type       string `toml:"type" json:"type"`
	LocalIP    string `toml:"localIP" json:"localIP"`
	LocalPort  int    `toml:"localPort" json:"localPort"`
	RemotePort int    `toml:"remotePort" json:"remotePort"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var (
	installDir  = "/usr/local/frp"
	configPath  = "/usr/local/frp/frpc.toml"
	frpcPath    = "/usr/local/frp/frpc"
	serviceName = "frpc.service"
	serviceFile = "/etc/systemd/system/frpc.service"
)

func main() {
	// 初始化检查和安装
	if err := initializeSystem(); err != nil {
		log.Fatalf("系统初始化失败: %v", err)
	}

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// 使用嵌入的静态文件
	log.Println("使用嵌入的 web 文件")

	// 创建子文件系统
	webSubFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("创建 web 子文件系统失败: %v", err)
	}

	// 静态文件服务
	r.StaticFS("/static", http.FS(webSubFS))

	// Favicon 路由
	r.GET("/favicon.ico", func(c *gin.Context) {
		// 尝试读取 SVG favicon（作为 ICO 的替代）
		data, err := webFS.ReadFile("web/favicon.svg")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", data)
	})

	r.GET("/favicon.svg", func(c *gin.Context) {
		data, err := webFS.ReadFile("web/favicon.svg")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", data)
	})

	// 首页路由
	r.GET("/", func(c *gin.Context) {
		data, err := webFS.ReadFile("web/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "无法读取 index.html: %v", err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// API 路由
	api := r.Group("/api")
	{
		api.GET("/config", getConfig)
		api.POST("/config", saveConfig)
		api.GET("/status", getStatus)
		api.GET("/check", checkConfigFile)
		api.GET("/logs", getLogs)
		api.GET("/service-status", getServiceStatus)
		api.GET("/system-status", getSystemStatus)
		api.POST("/install-frpc", installFrpc)
		api.POST("/update-frpc", updateFrpc)
	}

	// 启动服务器
	fmt.Println("FRP 配置管理服务启动中...")
	fmt.Println("请在浏览器中访问: http://localhost:8888")

	// 自动打开浏览器
	go func() {
		time.Sleep(2 * time.Second)
		openBrowser("http://localhost:8888")
	}()

	log.Fatal(r.Run(":8888"))
}

// 系统初始化检查
func initializeSystem() error {
	log.Println("开始系统初始化检查...")

	// 1. 检查并创建安装目录
	if err := ensureInstallDir(); err != nil {
		return fmt.Errorf("创建安装目录失败: %v", err)
	}

	// 2. 检查 frpc 客户端是否存在
	frpcExists := checkFrpcExists()
	if !frpcExists {
		log.Println("frpc 客户端不存在，需要安装")
		return nil // 不在启动时安装，让用户通过 Web 界面安装
	}

	// 3. 检查系统服务是否存在
	serviceExists := checkServiceExists()
	if !serviceExists {
		log.Println("系统服务不存在，需要创建")
		if err := createSystemService(); err != nil {
			log.Printf("创建系统服务失败: %v", err)
		}
	}

	// 4. 启动服务（如果配置文件存在）
	if _, err := os.Stat(configPath); err == nil {
		log.Println("配置文件存在，尝试启动服务...")
		if err := startSystemService(); err != nil {
			log.Printf("启动服务失败: %v", err)
		}
	}

	log.Println("系统初始化检查完成")
	return nil
}

// 确保安装目录存在
func ensureInstallDir() error {
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		log.Printf("创建安装目录: %s", installDir)
		return os.MkdirAll(installDir, 0755)
	}
	return nil
}

// 检查 frpc 客户端是否存在
func checkFrpcExists() bool {
	_, err := os.Stat(frpcPath)
	return err == nil
}

// 检查系统服务是否存在
func checkServiceExists() bool {
	_, err := os.Stat(serviceFile)
	return err == nil
}

// 创建系统服务
func createSystemService() error {
	serviceContent := fmt.Sprintf(`[Unit]
Description=FRP Client
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Restart=on-failure
RestartSec=5s
ExecStart=%s -c %s
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
`, frpcPath, configPath)

	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return err
	}

	// 重新加载 systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return err
	}

	// 启用服务
	if err := exec.Command("systemctl", "enable", serviceName).Run(); err != nil {
		return err
	}

	log.Println("系统服务创建成功")
	return nil
}

// 启动系统服务
func startSystemService() error {
	return exec.Command("systemctl", "start", serviceName).Run()
}

// 检查配置文件是否存在
func checkConfigFile(c *gin.Context) {
	_, err := os.Stat(configPath)
	exists := !os.IsNotExist(err)

	c.JSON(http.StatusOK, gin.H{
		"exists": exists,
		"path":   configPath,
	})
}

// 获取配置
func getConfig(c *gin.Context) {
	config, err := loadConfigFromFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// 保存配置并重启服务
func saveConfig(c *gin.Context) {
	var config Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的配置数据: " + err.Error()})
		return
	}

	// 保存配置到文件
	if err := saveConfigToFile(&config); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "保存配置失败: " + err.Error()})
		return
	}

	// 检查 frpc 是否已安装并且服务正常
	frpcPath := "/usr/local/frp/frpc"
	if _, err := os.Stat(frpcPath); os.IsNotExist(err) {
		// frpc 未安装，返回成功但提示需要安装
		c.JSON(http.StatusOK, gin.H{
			"message":     "配置已保存，请先安装 FRP 客户端",
			"needInstall": true,
		})
		return
	}

	// 检查 systemctl 服务是否存在并可用
	checkCmd := exec.Command("systemctl", "is-enabled", "frpc.service")
	if err := checkCmd.Run(); err != nil {
		// 服务未正确安装，返回成功但提示需要安装
		c.JSON(http.StatusOK, gin.H{
			"message":     "配置已保存，请先完成 FRP 客户端安装",
			"needInstall": true,
		})
		return
	}

	// 重启 frpc 服务
	if err := restartFrpcService(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "保存配置失败: 重启服务失败: " + err.Error()})
		return
	}

	// 检查服务连接状态
	connectionStatus := checkServerConnection()

	response := gin.H{
		"message":          "配置保存成功，服务已重启",
		"connectionStatus": connectionStatus,
	}

	c.JSON(http.StatusOK, response)
}

// 获取服务状态
func getStatus(c *gin.Context) {
	// 使用 systemctl 检查服务状态
	statusCmd := exec.Command("systemctl", "is-active", "frpc.service")
	output, err := statusCmd.Output()

	status := "stopped"
	if err == nil {
		statusStr := strings.TrimSpace(string(output))
		if statusStr == "active" {
			status = "running"
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// 获取服务日志
func getLogs(c *gin.Context) {
	cmd := exec.Command("journalctl", "-u", "frpc.service", "--no-pager", "-n", "50")
	output, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取日志失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": string(output)})
}

// 获取 systemctl 服务状态
func getServiceStatus(c *gin.Context) {
	// 检查服务状态
	statusCmd := exec.Command("systemctl", "is-active", "frpc.service")
	statusOutput, _ := statusCmd.Output()
	status := string(statusOutput)
	status = strings.TrimSpace(status)

	// 获取服务详细信息
	infoCmd := exec.Command("systemctl", "status", "frpc.service", "--no-pager", "-l")
	infoOutput, _ := infoCmd.Output()

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"info":   string(infoOutput),
	})
}

// 获取系统状态
func getSystemStatus(c *gin.Context) {
	status := map[string]interface{}{
		"installDirExists": checkInstallDirExists(),
		"frpcExists":       checkFrpcExists(),
		"serviceExists":    checkServiceExists(),
		"configExists":     checkConfigExists(),
		"frpcVersion":      getFrpcVersion(),
		"latestVersion":    getLatestVersion(),
	}

	c.JSON(http.StatusOK, status)
}

// 安装 frpc
func installFrpc(c *gin.Context) {
	if err := downloadAndInstallFrpc(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "安装失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "frpc 安装成功"})
}

// 更新 frpc
func updateFrpc(c *gin.Context) {
	if err := downloadAndInstallFrpc(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "frpc 更新成功"})
}

// 从文件加载配置
func loadConfigFromFile() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// 保存配置到文件
func saveConfigToFile(config *Config) error {
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 备份原配置文件
	backupPath := configPath + ".backup." + time.Now().Format("20060102150405")
	if _, err := os.Stat(configPath); err == nil {
		if err := copyFile(configPath, backupPath); err != nil {
			log.Printf("备份配置文件失败: %v", err)
		}
	}

	// 写入新配置
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// 重启 frpc 服务
func restartFrpcService() error {
	// 使用 systemctl 重启服务
	cmd := exec.Command("systemctl", "restart", "frpc.service")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("重启 frpc 服务失败: %v", err)
	}

	log.Println("frpc 服务已通过 systemctl 重启")

	// 等待服务启动，最多等待10秒
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)

		// 检查服务状态
		statusCmd := exec.Command("systemctl", "is-active", "frpc.service")
		output, err := statusCmd.Output()
		if err != nil {
			continue // 继续等待
		}

		status := strings.TrimSpace(string(output))
		if status == "active" {
			log.Printf("frpc 服务状态: %s", status)
			return nil
		} else if status == "activating" {
			log.Printf("frpc 服务正在启动中: %s", status)
			continue // 继续等待
		} else if status == "failed" {
			return fmt.Errorf("服务启动失败: %s", status)
		}
	}

	// 最后一次检查
	statusCmd := exec.Command("systemctl", "is-active", "frpc.service")
	output, err := statusCmd.Output()
	if err != nil {
		log.Printf("警告: 无法检查服务状态，但重启命令已执行")
		return nil // 不返回错误，因为重启命令已经执行
	}

	status := strings.TrimSpace(string(output))
	log.Printf("frpc 服务最终状态: %s", status)

	// 只要不是 failed 状态，就认为成功
	if status == "failed" {
		return fmt.Errorf("服务启动失败: %s", status)
	}

	return nil
}

// 检查服务器连接状态
func checkServerConnection() map[string]interface{} {
	// 等待服务完全启动
	time.Sleep(3 * time.Second)

	// 获取最近的日志来检查连接状态
	cmd := exec.Command("journalctl", "-u", "frpc.service", "--no-pager", "-n", "20", "--since", "30 seconds ago")
	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"connected": false,
			"error":     "无法获取服务日志: " + err.Error(),
		}
	}

	logs := string(output)

	// 检查日志中的连接状态
	if strings.Contains(logs, "login to server success") {
		return map[string]interface{}{
			"connected": true,
			"message":   "成功连接到服务器",
			"logs":      logs,
		}
	} else if strings.Contains(logs, "connect to server error") || strings.Contains(logs, "login to the server failed") {
		return map[string]interface{}{
			"connected": false,
			"message":   "连接服务器失败",
			"logs":      logs,
		}
	} else {
		return map[string]interface{}{
			"connected": false,
			"message":   "服务状态未知，请查看日志",
			"logs":      logs,
		}
	}
}

// 复制文件
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// 打开浏览器
func openBrowser(url string) {
	var cmd string
	var args []string

	switch {
	case isCommandAvailable("xdg-open"):
		cmd = "xdg-open"
		args = []string{url}
	case isCommandAvailable("open"):
		cmd = "open"
		args = []string{url}
	case isCommandAvailable("cmd"):
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		log.Printf("无法自动打开浏览器，请手动访问: %s", url)
		return
	}

	if err := exec.Command(cmd, args...).Start(); err != nil {
		log.Printf("打开浏览器失败: %v", err)
		log.Printf("请手动访问: %s", url)
	}
}

// 检查命令是否可用
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// 检查安装目录是否存在
func checkInstallDirExists() bool {
	_, err := os.Stat(installDir)
	return err == nil
}

// 检查配置文件是否存在
func checkConfigExists() bool {
	_, err := os.Stat(configPath)
	return err == nil
}

// 获取 frpc 版本
func getFrpcVersion() string {
	if !checkFrpcExists() {
		return ""
	}

	cmd := exec.Command(frpcPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// 解析版本号
	version := strings.TrimSpace(string(output))
	if strings.Contains(version, "frpc version") {
		parts := strings.Fields(version)
		if len(parts) >= 3 {
			return parts[2]
		}
	}

	return version
}

// 获取最新版本
func getLatestVersion() string {
	// 尝试多个源获取版本信息
	sources := []string{
		"https://ghfast.top/https://github.com/fatedier/frp/releases/latest",
		"https://hk.gh-proxy.com/https://github.com/fatedier/frp/releases/latest",
		"https://gh-proxy.com/https://github.com/fatedier/frp/releases/latest",
		"https://hk.gh-proxy.com/https://github.com/fatedier/frp/releases/latest",
		"https://gh-proxy.com/https://github.com/fatedier/frp/releases/latest",
		"https://cdn.gh-proxy.com/https://github.com/fatedier/frp/releases/latest",
		"https://edgeone.gh-proxy.com/https://github.com/fatedier/frp/releases/latest",
	}

	for _, source := range sources {
		log.Printf("尝试从 %s 获取版本信息...", source)
		cmd := exec.Command("curl", "-sL", "--connect-timeout", "3", source)
		output, err := cmd.Output()
		if err != nil {
			log.Printf("从 %s 获取失败: %v", source, err)
			continue
		}

		content := string(output)

		// 简单的版本提取逻辑
		if strings.Contains(content, "/frp/releases/tag/v") {
			start := strings.Index(content, "/frp/releases/tag/v")
			if start != -1 {
				start += len("/frp/releases/tag/v")
				end := start
				for end < len(content) && (content[end] >= '0' && content[end] <= '9' || content[end] == '.') {
					end++
				}
				if end > start {
					version := content[start:end]
					log.Printf("获取到版本: %s", version)
					return version
				}
			}
		}
	}

	// 如果都失败了，返回一个已知的稳定版本
	log.Println("无法获取最新版本，使用默认版本 0.63.0")
	return "0.63.0"
}

// 下载并安装 frpc
func downloadAndInstallFrpc() error {
	log.Println("开始下载并安装 frpc...")

	// 确保安装目录存在
	if err := ensureInstallDir(); err != nil {
		return err
	}

	// 获取最新版本
	version := getLatestVersion()
	if version == "" {
		return fmt.Errorf("无法获取最新版本号")
	}

	log.Printf("最新版本: %s", version)

	// 构造下载链接（多个源）
	downloadURLs := []string{
		fmt.Sprintf("https://github.com/fatedier/frp/releases/download/v%s/frp_%s_linux_amd64.tar.gz", version, version),
		fmt.Sprintf("https://ghfast.top/https://github.com/fatedier/frp/releases/download/v%s/frp_%s_linux_amd64.tar.gz", version, version),
	}
	downloadFile := fmt.Sprintf("/tmp/frp_%s_linux_amd64.tar.gz", version)

	// 尝试从多个源下载文件
	var downloadErr error
	for _, downloadURL := range downloadURLs {
		log.Printf("尝试下载: %s", downloadURL)
		cmd := exec.Command("curl", "-L", "--connect-timeout", "30", "--max-time", "300", "-o", downloadFile, downloadURL)
		if err := cmd.Run(); err != nil {
			downloadErr = err
			log.Printf("下载失败: %v", err)
			continue
		}

		// 检查文件是否下载成功
		if _, err := os.Stat(downloadFile); err == nil {
			log.Println("下载成功")
			downloadErr = nil
			break
		}
	}

	if downloadErr != nil {
		return fmt.Errorf("所有下载源都失败: %v", downloadErr)
	}

	// 验证文件
	verifyCmd := exec.Command("tar", "-tzf", downloadFile)
	if err := verifyCmd.Run(); err != nil {
		os.Remove(downloadFile)
		return fmt.Errorf("下载的文件无效: %v", err)
	}

	// 解压文件
	log.Println("解压文件...")
	extractCmd := exec.Command("tar", "-xzf", downloadFile, "-C", "/tmp/")
	if err := extractCmd.Run(); err != nil {
		os.Remove(downloadFile)
		return fmt.Errorf("解压失败: %v", err)
	}

	// 复制 frpc 到安装目录
	srcPath := fmt.Sprintf("/tmp/frp_%s_linux_amd64/frpc", version)
	if err := copyFile(srcPath, frpcPath); err != nil {
		return fmt.Errorf("复制文件失败: %v", err)
	}

	// 设置可执行权限
	if err := os.Chmod(frpcPath, 0755); err != nil {
		return fmt.Errorf("设置权限失败: %v", err)
	}

	// 创建系统服务（如果不存在）
	if !checkServiceExists() {
		if err := createSystemService(); err != nil {
			log.Printf("创建系统服务失败: %v", err)
		}
	}

	// 清理临时文件
	os.Remove(downloadFile)
	os.RemoveAll(fmt.Sprintf("/tmp/frp_%s_linux_amd64", version))

	log.Println("frpc 安装完成")
	return nil
}
