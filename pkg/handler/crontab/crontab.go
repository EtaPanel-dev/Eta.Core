package crontab

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models"
	"github.com/gin-gonic/gin"
)

// GetCrontabList 获取crontab列表
// @Summary 获取crontab列表
// @Description 获取系统中所有的crontab定时任务
// @Tags 定时任务
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=[]models.CrontabEntry} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/crontab [get]
func GetCrontabList(c *gin.Context) {
	entries, err := getCrontabEntries()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取crontab列表失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "获取crontab列表成功", entries)
}

// CreateCrontabEntry 创建crontab条目
// @Summary 创建crontab条目
// @Description 创建新的定时任务
// @Tags 定时任务
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CrontabRequest true "定时任务信息"
// @Success 200 {object} handler.Response "创建成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/crontab [post]
func CreateCrontabEntry(c *gin.Context) {
	var req models.CrontabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	// 验证cron表达式
	if err := validateCronExpression(req.Minute, req.Hour, req.Day, req.Month, req.Weekday); err != nil {
		handler.Respond(c, http.StatusBadRequest, "cron表达式无效: "+err.Error(), nil)
		return
	}

	// 验证命令
	if strings.TrimSpace(req.Command) == "" {
		handler.Respond(c, http.StatusBadRequest, "命令不能为空", nil)
		return
	}

	// 获取当前crontab
	entries, err := getCrontabEntries()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取当前crontab失败: "+err.Error(), nil)
		return
	}

	// 创建新条目
	newEntry := models.CrontabEntry{
		Id:      getNextID(entries),
		Minute:  req.Minute,
		Hour:    req.Hour,
		Day:     req.Day,
		Month:   req.Month,
		Weekday: req.Weekday,
		Command: req.Command,
		Comment: req.Comment,
		Enabled: req.Enabled,
	}

	// 添加到列表
	entries = append(entries, newEntry)

	// 保存crontab
	if err := saveCrontab(entries); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "保存crontab失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "创建crontab条目成功", newEntry)
}

// UpdateCrontabEntry 更新crontab条目
// @Summary 更新crontab条目
// @Description 更新指定ID的定时任务
// @Tags 定时任务
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务ID"
// @Param request body models.CrontabRequest true "定时任务信息"
// @Success 200 {object} handler.Response "更新成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 404 {object} handler.Response "任务不存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/crontab/{id} [put]
func UpdateCrontabEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handler.Respond(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	var req models.CrontabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	// 验证cron表达式
	if err := validateCronExpression(req.Minute, req.Hour, req.Day, req.Month, req.Weekday); err != nil {
		handler.Respond(c, http.StatusBadRequest, "cron表达式无效: "+err.Error(), nil)
		return
	}

	// 验证命令
	if strings.TrimSpace(req.Command) == "" {
		handler.Respond(c, http.StatusBadRequest, "命令不能为空", nil)
		return
	}

	// 获取当前crontab
	entries, err := getCrontabEntries()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取当前crontab失败: "+err.Error(), nil)
		return
	}

	// 查找并更新条目
	found := false
	for i, entry := range entries {
		if entry.Id == id {
			entries[i].Minute = req.Minute
			entries[i].Hour = req.Hour
			entries[i].Day = req.Day
			entries[i].Month = req.Month
			entries[i].Weekday = req.Weekday
			entries[i].Command = req.Command
			entries[i].Comment = req.Comment
			entries[i].Enabled = req.Enabled
			found = true
			break
		}
	}

	if !found {
		handler.Respond(c, http.StatusNotFound, "未找到指定的crontab条目", nil)
		return
	}

	// 保存crontab
	if err := saveCrontab(entries); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "保存crontab失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "更新crontab条目成功", nil)
}

// DeleteCrontabEntry 删除crontab条目
// @Summary 删除crontab条目
// @Description 删除指定ID的定时任务
// @Tags 定时任务
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务ID"
// @Success 200 {object} handler.Response "删除成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 404 {object} handler.Response "任务不存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/crontab/{id} [delete]
func DeleteCrontabEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handler.Respond(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	// 获取当前crontab
	entries, err := getCrontabEntries()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取当前crontab失败: "+err.Error(), nil)
		return
	}

	// 查找并删除条目
	found := false
	newEntries := make([]models.CrontabEntry, 0)
	for _, entry := range entries {
		if entry.Id != id {
			newEntries = append(newEntries, entry)
		} else {
			found = true
		}
	}

	if !found {
		handler.Respond(c, http.StatusNotFound, "未找到指定的crontab条目", nil)
		return
	}

	// 保存crontab
	if err := saveCrontab(newEntries); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "保存crontab失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "删除crontab条目成功", nil)
}

// ToggleCrontabEntry 启用/禁用crontab条目
// @Summary 启用/禁用crontab条目
// @Description 切换指定ID定时任务的启用状态
// @Tags 定时任务
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务ID"
// @Success 200 {object} handler.Response "操作成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 404 {object} handler.Response "任务不存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/crontab/{id}/toggle [post]
func ToggleCrontabEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handler.Respond(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	// 获取当前crontab
	entries, err := getCrontabEntries()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取当前crontab失败: "+err.Error(), nil)
		return
	}

	// 查找并切换状态
	found := false
	for i, entry := range entries {
		if entry.Id == id {
			entries[i].Enabled = !entries[i].Enabled
			found = true
			break
		}
	}

	if !found {
		handler.Respond(c, http.StatusNotFound, "未找到指定的crontab条目", nil)
		return
	}

	// 保存crontab
	if err := saveCrontab(entries); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "保存crontab失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "切换crontab条目状态成功", nil)
}

// getCrontabEntries 获取crontab条目
func getCrontabEntries() ([]models.CrontabEntry, error) {
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.Output()
	if err != nil {
		// 如果没有crontab，返回空列表
		if strings.Contains(err.Error(), "no crontab") {
			return []models.CrontabEntry{}, nil
		}
		return nil, err
	}

	var entries []models.CrontabEntry
	lines := strings.Split(string(output), "\n")
	id := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		entry := parseCrontabLine(line, id)
		if entry != nil {
			entries = append(entries, *entry)
			id++
		}
	}

	return entries, nil
}

// parseCrontabLine 解析crontab行
func parseCrontabLine(line string, id int) *models.CrontabEntry {
	entry := &models.CrontabEntry{
		Id:       id,
		FullLine: line,
		Enabled:  true,
	}

	// 检查是否被注释（禁用）
	if strings.HasPrefix(line, "#") {
		entry.Enabled = false
		line = strings.TrimPrefix(line, "#")
		line = strings.TrimSpace(line)
	}

	// 提取注释
	commentIndex := strings.Index(line, "#")
	if commentIndex > 0 {
		entry.Comment = strings.TrimSpace(line[commentIndex+1:])
		line = strings.TrimSpace(line[:commentIndex])
	}

	// 解析cron字段
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return entry // 返回原始行，可能是注释或其他内容
	}

	entry.Minute = fields[0]
	entry.Hour = fields[1]
	entry.Day = fields[2]
	entry.Month = fields[3]
	entry.Weekday = fields[4]
	entry.Command = strings.Join(fields[5:], " ")

	// 计算下次运行时间
	entry.NextRun = calculateNextRun(entry.Minute, entry.Hour, entry.Day, entry.Month, entry.Weekday)

	return entry
}

// saveCrontab 保存crontab
func saveCrontab(entries []models.CrontabEntry) error {
	var lines []string

	for _, entry := range entries {
		line := fmt.Sprintf("%s %s %s %s %s %s",
			entry.Minute, entry.Hour, entry.Day, entry.Month, entry.Weekday, entry.Command)

		if entry.Comment != "" {
			line += " # " + entry.Comment
		}

		if !entry.Enabled {
			line = "# " + line
		}

		lines = append(lines, line)
	}

	// 写入临时文件
	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}

	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(content)

	return cmd.Run()
}

// validateCronExpression 验证cron表达式
func validateCronExpression(minute, hour, day, month, weekday string) error {
	// 验证分钟 (0-59)
	if err := validateCronField(minute, 0, 59, "分钟"); err != nil {
		return err
	}

	// 验证小时 (0-23)
	if err := validateCronField(hour, 0, 23, "小时"); err != nil {
		return err
	}

	// 验证日期 (1-31)
	if err := validateCronField(day, 1, 31, "日期"); err != nil {
		return err
	}

	// 验证月份 (1-12)
	if err := validateCronField(month, 1, 12, "月份"); err != nil {
		return err
	}

	// 验证星期 (0-7, 0和7都表示星期日)
	if err := validateCronField(weekday, 0, 7, "星期"); err != nil {
		return err
	}

	return nil
}

// validateCronField 验证cron字段
func validateCronField(field string, min, max int, fieldName string) error {
	if field == "*" {
		return nil
	}

	// 处理步长 (*/n)
	if strings.Contains(field, "/") {
		parts := strings.Split(field, "/")
		if len(parts) != 2 {
			return fmt.Errorf("%s字段格式错误", fieldName)
		}

		if parts[0] != "*" {
			if err := validateCronField(parts[0], min, max, fieldName); err != nil {
				return err
			}
		}

		step, err := strconv.Atoi(parts[1])
		if err != nil || step <= 0 {
			return fmt.Errorf("%s字段步长无效", fieldName)
		}

		return nil
	}

	// 处理范围 (n-m)
	if strings.Contains(field, "-") {
		parts := strings.Split(field, "-")
		if len(parts) != 2 {
			return fmt.Errorf("%s字段范围格式错误", fieldName)
		}

		start, err1 := strconv.Atoi(parts[0])
		end, err2 := strconv.Atoi(parts[1])

		if err1 != nil || err2 != nil {
			return fmt.Errorf("%s字段范围值无效", fieldName)
		}

		if start < min || start > max || end < min || end > max || start > end {
			return fmt.Errorf("%s字段范围超出有效范围 (%d-%d)", fieldName, min, max)
		}

		return nil
	}

	// 处理列表 (n,m,...)
	if strings.Contains(field, ",") {
		values := strings.Split(field, ",")
		for _, value := range values {
			if err := validateCronField(strings.TrimSpace(value), min, max, fieldName); err != nil {
				return err
			}
		}
		return nil
	}

	// 处理单个数值
	value, err := strconv.Atoi(field)
	if err != nil {
		return fmt.Errorf("%s字段值无效", fieldName)
	}

	if value < min || value > max {
		return fmt.Errorf("%s字段值超出有效范围 (%d-%d)", fieldName, min, max)
	}

	return nil
}

// calculateNextRun 计算下次运行时间（简化版本）
func calculateNextRun(minute, hour, day, month, weekday string) string {
	// 这是一个简化的实现，实际的cron计算会更复杂
	now := time.Now()

	// 如果是简单的数值，尝试计算下次运行时间
	if m, err := strconv.Atoi(minute); err == nil {
		if h, err := strconv.Atoi(hour); err == nil {
			next := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())
			if next.Before(now) {
				next = next.Add(24 * time.Hour)
			}
			return next.Format("2006-01-02 15:04:05")
		}
	}

	return "计算中..."
}

// getNextID 获取下一个ID
func getNextID(entries []models.CrontabEntry) int {
	maxID := 0
	for _, entry := range entries {
		if entry.Id > maxID {
			maxID = entry.Id
		}
	}
	return maxID + 1
}
