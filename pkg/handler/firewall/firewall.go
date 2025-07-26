package firewall

import (
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/gin-gonic/gin"
)

// FirewallRule 防火墙规则
type FirewallRule struct {
	ID       int    `json:"id"`
	Action   string `json:"action"`   // ALLOW, DENY, REJECT
	From     string `json:"from"`     // 来源IP/网段
	To       string `json:"to"`       // 目标端口/服务
	Protocol string `json:"protocol"` // tcp, udp, any
	Comment  string `json:"comment"`
}

// FirewallStatus 防火墙状态
type FirewallStatus struct {
	Active bool           `json:"active"`
	Rules  []FirewallRule `json:"rules"`
}

// GetFirewallStatus 获取防火墙状态
func GetFirewallStatus(c *gin.Context) {
	// 检查ufw状态
	cmd := exec.Command("ufw", "status", "numbered")
	output, err := cmd.Output()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取防火墙状态失败: "+err.Error(), nil)
		return
	}

	status := parseFirewallStatus(string(output))
	handler.Respond(c, http.StatusOK, "获取防火墙状态成功", status)
}

// EnableFirewall 启用防火墙
func EnableFirewall(c *gin.Context) {
	cmd := exec.Command("ufw", "--force", "enable")
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "启用防火墙失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "防火墙已启用", map[string]string{"output": string(output)})
}

// DisableFirewall 禁用防火墙
func DisableFirewall(c *gin.Context) {
	cmd := exec.Command("ufw", "disable")
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "禁用防火墙失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "防火墙已禁用", map[string]string{"output": string(output)})
}

// AddFirewallRule 添加防火墙规则
func AddFirewallRule(c *gin.Context) {
	var req struct {
		Action   string `json:"action" binding:"required"` // allow, deny, reject
		Port     string `json:"port"`                      // 端口号或端口范围
		Protocol string `json:"protocol"`                  // tcp, udp, any
		From     string `json:"from"`                      // 来源IP/网段
		Comment  string `json:"comment"`                   // 注释
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	args := []string{req.Action}

	if req.From != "" {
		args = append(args, "from", req.From)
	}

	if req.Port != "" {
		if req.Protocol != "" && req.Protocol != "any" {
			args = append(args, "to", "any", "port", req.Port, "proto", req.Protocol)
		} else {
			args = append(args, "to", "any", "port", req.Port)
		}
	}

	if req.Comment != "" {
		args = append(args, "comment", req.Comment)
	}

	cmd := exec.Command("ufw", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "添加防火墙规则失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "防火墙规则已添加", map[string]string{"output": string(output)})
}

// DeleteFirewallRule 删除防火墙规则
func DeleteFirewallRule(c *gin.Context) {
	ruleIDStr := c.Param("id")
	ruleID, err := strconv.Atoi(ruleIDStr)
	if err != nil {
		handler.Respond(c, http.StatusBadRequest, "无效的规则ID", nil)
		return
	}

	cmd := exec.Command("ufw", "--force", "delete", strconv.Itoa(ruleID))
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "删除防火墙规则失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "防火墙规则已删除", map[string]string{"output": string(output)})
}

// ResetFirewall 重置防火墙规则
func ResetFirewall(c *gin.Context) {
	cmd := exec.Command("ufw", "--force", "reset")
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "重置防火墙失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "防火墙已重置", map[string]string{"output": string(output)})
}

// AllowSSH 允许SSH连接（安全预设）
func AllowSSH(c *gin.Context) {
	cmd := exec.Command("ufw", "allow", "ssh")
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "允许SSH失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "SSH访问已允许", map[string]string{"output": string(output)})
}

// DenyAll 拒绝所有入站连接（安全预设）
func DenyAll(c *gin.Context) {
	cmd := exec.Command("ufw", "default", "deny", "incoming")
	output, err := cmd.CombinedOutput()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "设置默认拒绝失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "默认拒绝入站已设置", map[string]string{"output": string(output)})
}

// parseFirewallStatus 解析防火墙状态输出
func parseFirewallStatus(output string) FirewallStatus {
	lines := strings.Split(output, "\n")
	status := FirewallStatus{
		Active: false,
		Rules:  []FirewallRule{},
	}

	// 检查状态
	if len(lines) > 0 && strings.Contains(lines[0], "active") {
		status.Active = true
	}

	// 解析规则
	ruleRegex := regexp.MustCompile(`^\[\s*(\d+)\]\s+(\w+)\s+(.+)$`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := ruleRegex.FindStringSubmatch(line); len(matches) == 4 {
			id, _ := strconv.Atoi(matches[1])
			rule := FirewallRule{
				ID:     id,
				Action: strings.ToUpper(matches[2]),
				To:     strings.TrimSpace(matches[3]),
			}

			// 进一步解析规则详情
			parts := strings.Fields(matches[3])
			if len(parts) >= 2 {
				rule.To = parts[0]
				if len(parts) > 2 && strings.Contains(line, "from") {
					fromIndex := -1
					for i, part := range parts {
						if part == "from" && i+1 < len(parts) {
							fromIndex = i + 1
							break
						}
					}
					if fromIndex > 0 && fromIndex < len(parts) {
						rule.From = parts[fromIndex]
					}
				}
			}

			status.Rules = append(status.Rules, rule)
		}
	}

	return status
}
