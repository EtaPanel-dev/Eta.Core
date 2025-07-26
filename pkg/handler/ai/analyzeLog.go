package ai

import (
	"encoding/json"
	"net/http"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/extend/ai"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/gin-gonic/gin"
)

// AnalyzeLogRequest 日志分析请求结构
type AnalyzeLogRequest struct {
	LogContent string `json:"logContent" binding:"required" example:"[ERROR] 2024-01-01 12:00:00 Database connection failed"` // 要分析的日志内容
}

// AnalyzeLog 分析日志
// @Summary 智能日志分析
// @Description 使用Kimi AI分析日志内容，识别错误模式、异常情况和潜在问题
// @Tags AI助手
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AnalyzeLogRequest true "日志分析请求"
// @Success 200 {object} handler.Response{data=[]map[string]string} "分析结果，包含问题识别、建议解决方案等"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/ai/log [post]
func AnalyzeLog(c *gin.Context) {
	var request AnalyzeLogRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.Respond(c, http.StatusBadRequest, err, nil)
		return
	}

	// 使用 Kimi AI 进行日志分析
	analyzeLogJSON := ai.NewChatWithKimi(ai.LogAnalyzer, request.LogContent)
	var analyzeLogResponse []map[string]string
	err := json.Unmarshal([]byte(analyzeLogJSON), &analyzeLogResponse)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, err, nil)
		return
	}
	handler.Respond(c, http.StatusOK, nil, analyzeLogResponse)
	return
}
