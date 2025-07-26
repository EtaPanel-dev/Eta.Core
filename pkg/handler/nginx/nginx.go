package nginx

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models"
	"github.com/gin-gonic/gin"
)

// RestartNginx 重启Nginx服务
// @Summary 重启Nginx服务
// @Description 重启Nginx服务
// @Tags Nginx管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response "重启成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/nginx/restart [post]
func RestartNginx(c *gin.Context) {
	if err := restartNginxService(); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "重启Nginx失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "重启Nginx成功", nil)
}

// ReloadNginx 重新加载Nginx配置
// @Summary 重新加载Nginx配置
// @Description 重新加载Nginx配置而不重启服务
// @Tags Nginx管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response "重新加载成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/nginx/reload [post]
func ReloadNginx(c *gin.Context) {
	if err := reloadNginxService(); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "重新加载Nginx配置失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "重新加载Nginx配置成功", nil)
}

// TestNginxConfig 测试Nginx配置
// @Summary 测试Nginx配置
// @Description 测试Nginx配置文件的语法是否正确
// @Tags Nginx管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response "配置测试通过"
// @Failure 400 {object} handler.Response "配置测试失败"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/nginx/test [post]
func TestNginxConfig(c *gin.Context) {
	if err := testNginxConfig(); err != nil {
		handler.Respond(c, http.StatusBadRequest, "Nginx配置测试失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "Nginx配置测试通过", nil)
}

// getNginxStatus 获取Nginx状态信息
func getNginxStatus() models.NginxStatus {
	status := models.NginxStatus{}

	// 检查Nginx是否运行
	cmd := exec.Command("systemctl", "is-active", "nginx")
	output, err := cmd.Output()
	status.Running = err == nil && strings.TrimSpace(string(output)) == "active"

	if status.Running {
		// 获取PID
		if pidData, err := ioutil.ReadFile("/run/nginx.pid"); err == nil {
			if pid, err := strconv.Atoi(strings.TrimSpace(string(pidData))); err == nil {
				status.PID = pid
			}
		}

		// 获取运行时间
		cmd = exec.Command("systemctl", "show", "nginx", "--property=ActiveEnterTimestamp")
		if output, err := cmd.Output(); err == nil {
			timestampLine := strings.TrimSpace(string(output))
			if parts := strings.Split(timestampLine, "="); len(parts) == 2 {
				if startTime, err := time.Parse("Mon 2006-01-02 15:04:05 MST", parts[1]); err == nil {
					status.Uptime = time.Since(startTime).String()
				}
			}
		}
	}

	// 获取Nginx版本
	cmd = exec.Command("nginx", "-v")
	if output, err := cmd.CombinedOutput(); err == nil {
		versionLine := string(output)
		if matches := regexp.MustCompile(`nginx/([0-9.]+)`).FindStringSubmatch(versionLine); len(matches) > 1 {
			status.Version = matches[1]
		}
	}

	// 测试配置
	status.ConfigTest = testNginxConfig() == nil

	return status
}

// getNginxMainConfig 获取Nginx主配置
func getNginxMainConfig() (models.NginxConfig, error) {
	config := models.NginxConfig{
		ConfigPath: models.NginxConfigPath,
	}

	data, err := ioutil.ReadFile(models.NginxConfigPath)
	if err != nil {
		return config, err
	}

	content := string(data)

	// 解析配置项
	if matches := regexp.MustCompile(`user\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
		config.User = strings.TrimSpace(matches[1])
	}

	if matches := regexp.MustCompile(`worker_processes\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
		config.WorkerProcess = strings.TrimSpace(matches[1])
	}

	if matches := regexp.MustCompile(`error_log\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
		config.ErrorLog = strings.TrimSpace(matches[1])
	}

	if matches := regexp.MustCompile(`access_log\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
		config.AccessLog = strings.TrimSpace(matches[1])
	}

	if matches := regexp.MustCompile(`pid\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
		config.PidFile = strings.TrimSpace(matches[1])
	}

	if matches := regexp.MustCompile(`worker_connections\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
		config.WorkerConn = strings.TrimSpace(matches[1])
	}

	config.Gzip = strings.Contains(content, "gzip on;")
	config.ServerTokens = !strings.Contains(content, "server_tokens off;")

	return config, nil
}

// updateNginxMainConfig 更新Nginx主配置
func updateNginxMainConfig(config models.NginxConfig) error {
	// 备份原配置
	backupPath := models.NginxConfigPath + ".backup." + time.Now().Format("20060102150405")
	if err := copyFile(models.NginxConfigPath, backupPath); err != nil {
		return fmt.Errorf("备份配置文件失败: %v", err)
	}

	// 生成新配置内容
	newConfig := generateNginxConfig(config)

	// 写入新配置
	if err := ioutil.WriteFile(models.NginxConfigPath, []byte(newConfig), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	// 测试配置
	if err := testNginxConfig(); err != nil {
		// 配置有误，恢复备份
		copyFile(backupPath, models.NginxConfigPath)
		return fmt.Errorf("配置测试失败，已恢复备份: %v", err)
	}

	return nil
}

// resetNginxToDefault 重置Nginx为默认配置
func resetNginxToDefault() error {
	// 备份当前配置
	backupPath := models.NginxConfigPath + ".backup." + time.Now().Format("20060102150405")
	if err := copyFile(models.NginxConfigPath, backupPath); err != nil {
		return fmt.Errorf("备份配置文件失败: %v", err)
	}

	// 写入默认配置
	if err := ioutil.WriteFile(models.NginxConfigPath, []byte(models.DefaultNginxConfig), 0644); err != nil {
		return fmt.Errorf("写入默认配置失败: %v", err)
	}

	// 确保目录存在
	os.MkdirAll(models.NginxSitesAvailable, 0755)
	os.MkdirAll(models.NginxSitesEnabled, 0755)
	os.MkdirAll("/etc/nginx/conf.d", 0755)

	return nil
}

// getNginxSites 获取所有网站配置
func getNginxSites() ([]models.NginxSite, error) {
	var sites []models.NginxSite

	// 确保目录存在
	if _, err := os.Stat(models.NginxSitesAvailable); os.IsNotExist(err) {
		os.MkdirAll(models.NginxSitesAvailable, 0755)
		return sites, nil
	}

	files, err := ioutil.ReadDir(models.NginxSitesAvailable)
	if err != nil {
		return nil, err
	}

	id := 1
	for _, file := range files {
		if file.IsDir() || file.Name() == "default" {
			continue
		}

		sitePath := filepath.Join(models.NginxSitesAvailable, file.Name())
		site, err := parseSiteConfig(sitePath, id)
		if err != nil {
			continue // 跳过解析失败的配置
		}

		// 检查是否启用
		enabledPath := filepath.Join(models.NginxSitesAvailable, file.Name())
		if _, err := os.Stat(enabledPath); err == nil {
			site.Enabled = true
		}

		sites = append(sites, *site)
		id++
	}

	return sites, nil
}

// parseSiteConfig 解析网站配置文件
func parseSiteConfig(configPath string, id int) (*models.NginxSite, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	site := &models.NginxSite{
		ID:         id,
		Name:       filepath.Base(configPath),
		ConfigPath: configPath,
		Index:      "index.html index.htm",
	}

	// 解析server_name
	if matches := regexp.MustCompile(`server_name\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
		domains := strings.Fields(strings.TrimSpace(matches[1]))
		if len(domains) > 0 {
			site.Domain = domains[0]
			if len(domains) > 1 {
				site.Aliases = domains[1:]
			}
		}
	}

	// 解析root
	if matches := regexp.MustCompile(`root\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
		site.Root = strings.TrimSpace(matches[1])
	}

	// 解析index
	if matches := regexp.MustCompile(`index\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
		site.Index = strings.TrimSpace(matches[1])
	}

	// 检查SSL
	site.SSL = strings.Contains(content, "ssl_certificate")
	if site.SSL {
		if matches := regexp.MustCompile(`ssl_certificate\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
			site.SSLCert = strings.TrimSpace(matches[1])
		}
		if matches := regexp.MustCompile(`ssl_certificate_key\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
			site.SSLKey = strings.TrimSpace(matches[1])
		}
	}

	// 检查强制HTTPS
	site.ForceHTTPS = strings.Contains(content, "return 301 https://")

	// 检查反向代理
	site.Proxy = strings.Contains(content, "proxy_pass")
	if site.Proxy {
		if matches := regexp.MustCompile(`proxy_pass\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
			site.ProxyPass = strings.TrimSpace(matches[1])
		}
	}

	// 解析日志
	if matches := regexp.MustCompile(`access_log\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
		site.AccessLog = strings.TrimSpace(matches[1])
	}
	if matches := regexp.MustCompile(`error_log\s+([^;]+);`).FindStringSubmatch(content); len(matches) > 1 {
		site.ErrorLog = strings.TrimSpace(matches[1])
	}

	// 获取文件时间
	if info, err := os.Stat(configPath); err == nil {
		site.CreatedAt = info.ModTime().Format("2006-01-02 15:04:05")
		site.UpdatedAt = info.ModTime().Format("2006-01-02 15:04:05")
	}

	return site, nil
}

// domainExists 检查域名是否已存在
func domainExists(domain string) (bool, error) {
	sites, err := getNginxSites()
	if err != nil {
		return false, err
	}

	for _, site := range sites {
		if site.Domain == domain {
			return true, nil
		}
		for _, alias := range site.Aliases {
			if alias == domain {
				return true, nil
			}
		}
	}

	return false, nil
}

// createNginxSite 创建网站配置
func createNginxSite(site models.NginxSite) error {
	// 设置默认值
	if site.Root == "" {
		site.Root = "/var/www/" + site.Name
	}
	if site.Index == "" {
		site.Index = "index.html index.htm"
	}
	if site.AccessLog == "" {
		site.AccessLog = "/var/log/nginx/" + site.Name + "_access.log"
	}
	if site.ErrorLog == "" {
		site.ErrorLog = "/var/log/nginx/" + site.Name + "_error.log"
	}

	// 生成配置内容
	config := generateSiteConfig(site)

	// 写入配置文件
	configPath := filepath.Join(models.NginxSitesAvailable, site.Name)
	if err := ioutil.WriteFile(configPath, []byte(config), 0644); err != nil {
		return err
	}

	// 创建网站根目录
	if err := os.MkdirAll(site.Root, 0755); err != nil {
		return err
	}

	// 创建默认首页
	indexPath := filepath.Join(site.Root, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		defaultIndex := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Welcome to %s</title>
</head>
<body>
    <h1>Welcome to %s</h1>
    <p>This is the default page for %s.</p>
    <p>Please replace this file with your own content.</p>
</body>
</html>`, site.Domain, site.Domain, site.Domain)

		ioutil.WriteFile(indexPath, []byte(defaultIndex), 0644)
	}

	// 如果启用，创建软链接
	if site.Enabled {
		enabledPath := filepath.Join(models.NginxSitesAvailable, site.Name)
		os.Remove(enabledPath) // 删除可能存在的旧链接
		if err := os.Symlink(configPath, enabledPath); err != nil {
			return err
		}
	}

	return nil
}

// updateNginxSite 更新网站配置
func updateNginxSite(site models.NginxSite) error {
	sites, err := getNginxSites()
	if err != nil {
		return err
	}

	// 查找现有站点
	var existingSite *models.NginxSite
	for _, s := range sites {
		if s.ID == site.ID {
			existingSite = &s
			break
		}
	}

	if existingSite == nil {
		return fmt.Errorf("网站不存在")
	}

	// 更新配置
	site.Name = existingSite.Name
	site.ConfigPath = existingSite.ConfigPath

	// 生成新配置
	config := generateSiteConfig(site)

	// 写入配置文件
	if err := ioutil.WriteFile(site.ConfigPath, []byte(config), 0644); err != nil {
		return err
	}

	// 处理启用/禁用状态
	enabledPath := filepath.Join(models.NginxSitesAvailable, site.Name)
	if site.Enabled {
		// 启用站点
		os.Remove(enabledPath)
		if err := os.Symlink(site.ConfigPath, enabledPath); err != nil {
			return err
		}
	} else {
		// 禁用站点
		os.Remove(enabledPath)
	}

	return nil
}

// deleteNginxSite 删除网站
func deleteNginxSite(id int) error {
	sites, err := getNginxSites()
	if err != nil {
		return err
	}

	var site *models.NginxSite
	for _, s := range sites {
		if s.ID == id {
			site = &s
			break
		}
	}

	if site == nil {
		return fmt.Errorf("网站不存在")
	}

	// 删除enabled链接
	enabledPath := filepath.Join(models.NginxSitesAvailable, site.Name)
	os.Remove(enabledPath)

	// 删除配置文件
	if err := os.Remove(site.ConfigPath); err != nil {
		return err
	}

	return nil
}

// toggleNginxSite 切换网站启用状态
func toggleNginxSite(id int) error {
	sites, err := getNginxSites()
	if err != nil {
		return err
	}

	var site *models.NginxSite
	for _, s := range sites {
		if s.ID == id {
			site = &s
			break
		}
	}

	if site == nil {
		return fmt.Errorf("网站不存在")
	}

	enabledPath := filepath.Join(models.NginxSitesAvailable, site.Name)

	if site.Enabled {
		// 禁用站点
		return os.Remove(enabledPath)
	} else {
		// 启用站点
		os.Remove(enabledPath) // 删除可能存在的旧链接
		return os.Symlink(site.ConfigPath, enabledPath)
	}
}

// generateNginxConfig 生成Nginx主配置
func generateNginxConfig(config models.NginxConfig) string {
	gzipStatus := "off"
	if config.Gzip {
		gzipStatus = "on"
	}

	serverTokens := "on"
	if !config.ServerTokens {
		serverTokens = "off"
	}

	return fmt.Sprintf(`user %s;
worker_processes %s;
pid %s;
include /etc/nginx/modules-enabled/*.conf;

events {
    worker_connections %s;
}

http {
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    server_tokens %s;

    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    ssl_protocols TLSv1 TLSv1.1 TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log %s main;
    error_log %s;

    gzip %s;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    include /etc/nginx/conf.d/*.conf;
    include /etc/nginx/sites-enabled/*;
}
`, config.User, config.WorkerProcess, config.PidFile, config.WorkerConn,
		serverTokens, config.AccessLog, config.ErrorLog, gzipStatus)
}

// generateSiteConfig 生成网站配置
func generateSiteConfig(site models.NginxSite) string {
	var config strings.Builder

	// HTTP服务器块
	if site.ForceHTTPS && site.SSL {
		config.WriteString(fmt.Sprintf(`server {
    listen 80;
    server_name %s`, site.Domain))

		for _, alias := range site.Aliases {
			config.WriteString(" " + alias)
		}

		config.WriteString(fmt.Sprintf(`;
    return 301 https://$server_name$request_uri;
}

`))
	}

	// HTTPS或主服务器块
	config.WriteString("server {\n")

	if site.SSL {
		config.WriteString("    listen 443 ssl http2;\n")
		if !site.ForceHTTPS {
			config.WriteString("    listen 80;\n")
		}
	} else {
		config.WriteString("    listen 80;\n")
	}

	config.WriteString(fmt.Sprintf("    server_name %s", site.Domain))
	for _, alias := range site.Aliases {
		config.WriteString(" " + alias)
	}
	config.WriteString(";\n\n")

	// SSL配置
	if site.SSL && site.SSLCert != "" && site.SSLKey != "" {
		config.WriteString(fmt.Sprintf(`    ssl_certificate %s;
    ssl_certificate_key %s;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_stapling on;
    ssl_stapling_verify on;

`, site.SSLCert, site.SSLKey))
	}

	// 根目录和索引
	if !site.Proxy {
		config.WriteString(fmt.Sprintf("    root %s;\n", site.Root))
		config.WriteString(fmt.Sprintf("    index %s;\n\n", site.Index))
	}

	// 日志配置
	if site.AccessLog != "" {
		config.WriteString(fmt.Sprintf("    access_log %s;\n", site.AccessLog))
	}
	if site.ErrorLog != "" {
		config.WriteString(fmt.Sprintf("    error_log %s;\n", site.ErrorLog))
	}
	config.WriteString("\n")

	// 反向代理配置
	if site.Proxy && site.ProxyPass != "" {
		config.WriteString(`    location / {
        proxy_pass ` + site.ProxyPass + `;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
`)
	} else {
		// 静态文件配置
		config.WriteString(`    location / {
        try_files $uri $uri/ =404;
    }

    location ~ \.(css|js|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
`)
	}

	// 伪静态规则
	if site.Rewrite != "" {
		config.WriteString("\n    # 伪静态规则\n")
		config.WriteString(site.Rewrite)
		config.WriteString("\n")
	}

	// 安全配置
	config.WriteString(`
    # 安全配置
    location ~ /\. {
        deny all;
    }
    
    location ~ ~$ {
        deny all;
    }
`)

	config.WriteString("}\n")

	return config.String()
}

// 服务管理函数
func restartNginxService() error {
	cmd := exec.Command("systemctl", "restart", "nginx")
	return cmd.Run()
}

func reloadNginxService() error {
	cmd := exec.Command("systemctl", "reload", "nginx")
	return cmd.Run()
}

func testNginxConfig() error {
	cmd := exec.Command("nginx", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("配置测试失败: %s", string(output))
	}
	return nil
}

// 辅助函数
func copyFile(src, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
}
