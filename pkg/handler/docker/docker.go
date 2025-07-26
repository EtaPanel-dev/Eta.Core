package docker

import (
	"bufio"
	"context"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// GetDockerImages 获取Docker镜像列表
func GetDockerImages(c *gin.Context) {
	cmd := exec.Command("docker", "images", "--format", "table {{.ID}}\t{{.Repository}}\t{{.Tag}}\t{{.CreatedAt}}\t{{.Size}}")
	output, err := cmd.Output()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取Docker镜像失败: "+err.Error(), nil)
		return
	}

	images := parseDockerImages(string(output))
	handler.Respond(c, http.StatusOK, "获取Docker镜像成功", images)
}

// GetDockerContainers 获取Docker容器列表
func GetDockerContainers(c *gin.Context) {
	showAll := c.Query("all") == "true"
	args := []string{"ps", "--format", "table {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}\t{{.CreatedAt}}"}
	if showAll {
		args = append(args, "-a")
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取Docker容器失败: "+err.Error(), nil)
		return
	}

	containers := parseDockerContainers(string(output))
	handler.Respond(c, http.StatusOK, "获取Docker容器成功", containers)
}

// PullDockerImage 拉取Docker镜像
func PullDockerImage(c *gin.Context) {
	var req struct {
		Image string `json:"image" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	cmd := exec.Command("docker", "pull", req.Image)
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "拉取镜像失败: "+err.Error(), string(output))
		return
	}

	handler.Respond(c, http.StatusOK, "拉取镜像成功", map[string]string{
		"image":  req.Image,
		"output": string(output),
	})
}

// DeleteDockerImage 删除Docker镜像
func DeleteDockerImage(c *gin.Context) {
	imageID := c.Param("id")
	if imageID == "" {
		handler.Respond(c, http.StatusBadRequest, "镜像ID不能为空", nil)
		return
	}

	force := c.Query("force") == "true"
	args := []string{"rmi"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, imageID)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "删除镜像失败: "+err.Error(), string(output))
		return
	}

	handler.Respond(c, http.StatusOK, "删除镜像成功", map[string]string{
		"image_id": imageID,
		"output":   string(output),
	})
}

// RunDockerContainer 运行Docker容器
func RunDockerContainer(c *gin.Context) {
	var req struct {
		Image       string            `json:"image" binding:"required"`
		Name        string            `json:"name"`
		Ports       map[string]string `json:"ports"`
		Volumes     map[string]string `json:"volumes"`
		Environment map[string]string `json:"environment"`
		Command     []string          `json:"command"`
		Detach      bool              `json:"detach"`
		Interactive bool              `json:"interactive"`
		TTY         bool              `json:"tty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	args := []string{"run"}

	if req.Detach {
		args = append(args, "-d")
	}
	if req.Interactive {
		args = append(args, "-i")
	}
	if req.TTY {
		args = append(args, "-t")
	}
	if req.Name != "" {
		args = append(args, "--name", req.Name)
	}

	// 添加端口映射
	for hostPort, containerPort := range req.Ports {
		args = append(args, "-p", hostPort+":"+containerPort)
	}

	// 添加卷挂载
	for hostPath, containerPath := range req.Volumes {
		args = append(args, "-v", hostPath+":"+containerPath)
	}

	// 添加环境变量
	for key, value := range req.Environment {
		args = append(args, "-e", key+"="+value)
	}

	args = append(args, req.Image)
	args = append(args, req.Command...)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "运行容器失败: "+err.Error(), string(output))
		return
	}

	handler.Respond(c, http.StatusOK, "运行容器成功", map[string]string{
		"container_id": strings.TrimSpace(string(output)),
		"output":       string(output),
	})
}

// StopDockerContainer 停止Docker容器
func StopDockerContainer(c *gin.Context) {
	containerID := c.Param("id")
	if containerID == "" {
		handler.Respond(c, http.StatusBadRequest, "容器ID不能为空", nil)
		return
	}

	cmd := exec.Command("docker", "stop", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "停止容器失败: "+err.Error(), string(output))
		return
	}

	handler.Respond(c, http.StatusOK, "停止容器成功", map[string]string{
		"container_id": containerID,
		"output":       string(output),
	})
}

// StartDockerContainer 启动Docker容器
func StartDockerContainer(c *gin.Context) {
	containerID := c.Param("id")
	if containerID == "" {
		handler.Respond(c, http.StatusBadRequest, "容器ID不能为空", nil)
		return
	}

	cmd := exec.Command("docker", "start", containerID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "启动容器失败: "+err.Error(), string(output))
		return
	}

	handler.Respond(c, http.StatusOK, "启动容器成功", map[string]string{
		"container_id": containerID,
		"output":       string(output),
	})
}

// RemoveDockerContainer 删除Docker容器
func RemoveDockerContainer(c *gin.Context) {
	containerID := c.Param("id")
	if containerID == "" {
		handler.Respond(c, http.StatusBadRequest, "容器ID不能为空", nil)
		return
	}

	force := c.Query("force") == "true"
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, containerID)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "删除容器失败: "+err.Error(), string(output))
		return
	}

	handler.Respond(c, http.StatusOK, "删除容器成功", map[string]string{
		"container_id": containerID,
		"output":       string(output),
	})
}

// SetDockerRegistry 设置Docker镜像源
func SetDockerRegistry(c *gin.Context) {
	var req models.DockerRegistry
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	// 这里可以实现镜像源配置逻辑
	// 例如修改 /etc/docker/daemon.json 文件
	handler.Respond(c, http.StatusOK, "设置镜像源成功", req)
}

// DockerTerminal WebSocket终端连接
func DockerTerminal(c *gin.Context) {
	containerID := c.Param("id")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "容器ID不能为空"})
		return
	}

	conn, err := models.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket升级失败: " + err.Error()})
		return
	}
	defer conn.Close()

	// 创建docker exec命令
	cmd := exec.Command("docker", "exec", "-it", containerID, "/bin/bash")

	// 如果bash不存在，尝试sh
	if err := exec.Command("docker", "exec", containerID, "which", "bash").Run(); err != nil {
		cmd = exec.Command("docker", "exec", "-it", containerID, "/bin/sh")
	}

	// 获取命令的stdin, stdout, stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "创建stdin失败: " + err.Error()})
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "创建stdout失败: " + err.Error()})
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "创建stderr失败: " + err.Error()})
		return
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		conn.WriteJSON(map[string]string{"error": "启动终端失败: " + err.Error()})
		return
	}

	// 创建context用于控制goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理WebSocket输入 -> Docker容器stdin
	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var msg map[string]string
				if err := conn.ReadJSON(&msg); err != nil {
					return
				}
				if input, ok := msg["input"]; ok {
					stdin.Write([]byte(input))
				}
			}
		}
	}()

	// 处理Docker容器stdout -> WebSocket输出
	go func() {
		defer cancel()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				conn.WriteJSON(map[string]string{"output": scanner.Text() + "\n"})
			}
		}
	}()

	// 处理Docker容器stderr -> WebSocket输出
	go func() {
		defer cancel()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				conn.WriteJSON(map[string]string{"error": scanner.Text() + "\n"})
			}
		}
	}()

	// 等待命令结束或context取消
	go func() {
		cmd.Wait()
		cancel()
	}()

	<-ctx.Done()
}

// 解析Docker镜像输出
func parseDockerImages(output string) []models.DockerImage {
	var images []models.DockerImage
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // 跳过标题行和空行
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			created, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", strings.Join(fields[3:len(fields)-1], " "))
			images = append(images, models.DockerImage{
				ID:         fields[0],
				Repository: fields[1],
				Tag:        fields[2],
				Created:    created,
				Size:       fields[len(fields)-1],
			})
		}
	}

	return images
}

// 解析Docker容器输出
func parseDockerContainers(output string) []models.DockerContainer {
	var containers []models.DockerContainer
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // 跳过标题行和空行
		}

		fields := strings.Fields(line)
		if len(fields) >= 6 {
			containers = append(containers, models.DockerContainer{
				ID:      fields[0],
				Name:    fields[1],
				Image:   fields[2],
				Status:  fields[3],
				Ports:   fields[4],
				Created: strings.Join(fields[5:], " "),
			})
		}
	}

	return containers
}
