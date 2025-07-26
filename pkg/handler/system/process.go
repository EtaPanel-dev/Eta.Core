package system

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models"
	"github.com/gin-gonic/gin"
)

// getProcessCount 获取进程数量
func getProcessCount() int {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "processes") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if count, err := strconv.Atoi(fields[1]); err == nil {
					return count
				}
			}
		}
	}

	// 如果无法从/proc/stat获取，则统计/proc目录下的进程数
	procDir, err := os.Open("/proc")
	if err != nil {
		return 0
	}
	defer procDir.Close()

	entries, err := procDir.Readdir(-1)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			if _, err := strconv.Atoi(entry.Name()); err == nil {
				count++
			}
		}
	}

	return count
}

// GetProcessList 获取进程列表
// @Summary 获取进程列表
// @Description 获取系统中所有运行的进程信息
// @Tags 系统监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=object{processes=[]models.ProcessInfo}} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/system/processes [get]
func GetProcessList(c *gin.Context) {
	processes := getProcessList()
	handler.Respond(c, http.StatusOK, nil, gin.H{"processes": processes})
}

// KillProcess 终止进程
// @Summary 终止进程
// @Description 向指定进程发送信号以终止进程
// @Tags 系统监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{pid=int,signal=string} true "进程ID和信号类型"
// @Success 200 {object} handler.Response "终止成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "无法终止系统关键进程"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/system/process/kill [post]
func KillProcess(c *gin.Context) {
	var req struct {
		PID    int    `json:"pid"`
		Signal string `json:"signal,omitempty"` // TERM, KILL, HUP, etc.
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, nil, gin.H{"error": "请求参数错误"})
		return
	}

	if req.PID <= 0 {
		handler.Respond(c, http.StatusBadRequest, nil, gin.H{"error": "无效的进程ID"})
		return
	}

	// 检查是否为系统关键进程
	if isSystemCriticalProcess(req.PID) {
		handler.Respond(c, http.StatusForbidden, nil, gin.H{"error": "无法终止系统关键进程"})
		return
	}

	// 默认使用TERM信号
	signal := req.Signal
	if signal == "" {
		signal = "TERM"
	}

	err := killProcess(req.PID, signal)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, nil, gin.H{"error": err.Error()})
		return
	}

	handler.Respond(c, http.StatusOK, fmt.Sprintf("进程 %d 已发送 %s 信号", req.PID, signal), nil)
}

// killProcess 终止进程
func killProcess(pid int, signal string) error {
	// 验证进程是否存在
	if _, err := os.Stat(fmt.Sprintf("/proc/%d", pid)); os.IsNotExist(err) {
		return fmt.Errorf("进程 %d 不存在", pid)
	}

	// 构建kill命令
	var cmd *exec.Cmd
	switch signal {
	case "TERM":
		cmd = exec.Command("kill", "-TERM", strconv.Itoa(pid))
	case "KILL":
		cmd = exec.Command("kill", "-KILL", strconv.Itoa(pid))
	case "HUP":
		cmd = exec.Command("kill", "-HUP", strconv.Itoa(pid))
	case "INT":
		cmd = exec.Command("kill", "-INT", strconv.Itoa(pid))
	case "QUIT":
		cmd = exec.Command("kill", "-QUIT", strconv.Itoa(pid))
	case "USR1":
		cmd = exec.Command("kill", "-USR1", strconv.Itoa(pid))
	case "USR2":
		cmd = exec.Command("kill", "-USR2", strconv.Itoa(pid))
	default:
		return fmt.Errorf("不支持的信号: %s", signal)
	}

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("终止进程失败: %s", string(output))
	}

	return nil
}

// isSystemCriticalProcess 检查是否为系统关键进程
func isSystemCriticalProcess(pid int) bool {
	// 保护的系统进程PID和名称
	criticalPIDs := []int{1} // init进程
	criticalNames := []string{
		"systemd",
		"kthreadd",
		"ksoftirqd",
		"migration",
		"rcu_",
		"watchdog",
		"sshd",
		"NetworkManager",
		"dbus",
	}

	// 检查PID
	for _, criticalPID := range criticalPIDs {
		if pid == criticalPID {
			return true
		}
	}

	// 检查进程名称
	processInfo := getProcessInfo(pid)
	if processInfo != nil {
		for _, criticalName := range criticalNames {
			if strings.Contains(processInfo.Name, criticalName) {
				return true
			}
		}
	}

	return false
}

// getProcessList 获取进程列表
func getProcessList() []models.ProcessInfo {
	var processes []models.ProcessInfo

	procDir, err := os.Open("/proc")
	if err != nil {
		return processes
	}
	defer procDir.Close()

	entries, err := procDir.Readdir(-1)
	if err != nil {
		return processes
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if pid, err := strconv.Atoi(entry.Name()); err == nil {
				if processInfo := getProcessInfo(pid); processInfo != nil {
					processes = append(processes, *processInfo)
				}
			}
		}
	}

	return processes
}

// getProcessInfo 获取单个进程信息
func getProcessInfo(pid int) *models.ProcessInfo {
	procPath := fmt.Sprintf("/proc/%d", pid)

	// 检查进程是否存在
	if _, err := os.Stat(procPath); os.IsNotExist(err) {
		return nil
	}

	process := &models.ProcessInfo{PId: pid}

	// 读取进程状态
	if data, err := os.ReadFile(procPath + "/stat"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 3 {
			process.Name = strings.Trim(fields[1], "()")
			process.State = fields[2]
		}
	}

	// 读取进程命令行
	if data, err := os.ReadFile(procPath + "/cmdline"); err == nil {
		cmdline := string(data)
		if cmdline != "" {
			cmdline = strings.ReplaceAll(cmdline, "\x00", " ")
			process.Command = strings.TrimSpace(cmdline)
		} else {
			// 如果cmdline为空，使用进程名
			process.Command = process.Name
		}
	}

	// 读取进程内存信息
	if data, err := os.ReadFile(procPath + "/status"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "VmRSS:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if mem, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
						process.Memory = mem * 1024 // 转换为字节
					}
				}
				break
			}
		}
	}

	// 如果进程名为空，设置为unknown
	if process.Name == "" {
		process.Name = "unknown"
	}

	return process
}
