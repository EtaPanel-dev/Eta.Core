package ai

import (
	"encoding/json"
	"net/http"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/extend/ai"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/gin-gonic/gin"
)

// AnalyzeFilesRequest 文件分析请求结构
type AnalyzeFilesRequest struct {
	Files string `json:"files" binding:"required" example:"[{'name': 'app.log', 'size': 1024, 'modified': '2024-01-01'}, {'name': 'temp.txt', 'size': 512, 'modified': '2023-12-01'}]"` // 要分析的文件列表JSON字符串
}

// AnalyzeFiles 智能文件分析
// @Summary 智能文件分析和清理建议
// @Description 使用Kimi AI分析文件列表，识别可清理的文件、重复文件和存储优化建议
// @Tags AI助手
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AnalyzeFilesRequest true "文件分析请求"
// @Success 200 {object} handler.Response{data=[]map[string]string} "分析结果，包含清理建议、文件分类等"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/ai/files [post]
func AnalyzeFiles(c *gin.Context) {
	var request AnalyzeFilesRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.Respond(c, http.StatusBadRequest, err, nil)
		return
	}

	// 使用 Kimi AI 进行文件分析
	analyzeLogJSON := ai.NewChatWithKimi(ai.DirCleanPrompt, request.Files)
	var analyzeLogResponse []map[string]string
	err := json.Unmarshal([]byte(analyzeLogJSON), &analyzeLogResponse)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, err, nil)
		return
	}
	handler.Respond(c, http.StatusOK, nil, analyzeLogResponse)
	return
}
