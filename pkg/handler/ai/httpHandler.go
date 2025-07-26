package ai

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 全局AI处理器实例
var globalAIHandler *AIHandler

// 初始化全局AI处理器
func init() {
	globalAIHandler = NewAIHandler()
}

// 包级别的处理器函数，供路由调用

// ProcessQuery 处理自然语言查询
// @Summary 自然语言数据库查询
// @Description 使用自然语言与SQLite数据库进行交互，AI会自动将自然语言转换为相应的数据库操作
// @Tags AI工具链
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body QueryRequest true "查询请求"
// @Success 200 {object} APIResponse{data=object{response=string}} "查询成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 500 {object} APIResponse "服务器内部错误"
// @Router /api/auth/ai/query [post]
func ProcessQuery(c *gin.Context) {
	globalAIHandler.ProcessQuery(c)
}

// ExecuteToolCalls 执行工具调用
// @Summary 直接执行数据库工具调用
// @Description 直接执行预定义的数据库操作工具，支持批量调用
// @Tags AI工具链
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ToolCallRequest true "工具调用请求"
// @Success 200 {object} APIResponse{data=[]ToolCallResult} "执行成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 500 {object} APIResponse "执行失败"
// @Router /api/auth/ai/execute [post]
func ExecuteToolCalls(c *gin.Context) {
	globalAIHandler.ExecuteToolCalls(c)
}

// GetAvailableTools 获取可用工具列表
// @Summary 获取可用的数据库工具列表
// @Description 获取所有可用的数据库操作工具及其详细信息
// @Tags AI工具链
// @Produce json
// @Security BearerAuth
// @Success 200 {object} APIResponse{data=object{tools=[]ToolDefinition,count=int}} "获取成功"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 500 {object} APIResponse "服务器内部错误"
// @Router /api/auth/ai/tools [get]
func GetAvailableTools(c *gin.Context) {
	globalAIHandler.GetAvailableTools(c)
}

// GetTool 获取特定工具信息
// @Summary 获取特定工具的详细信息
// @Description 根据工具名称获取特定数据库工具的详细配置和参数信息
// @Tags AI工具链
// @Produce json
// @Security BearerAuth
// @Param name path string true "工具名称"
// @Success 200 {object} APIResponse{data=ToolDefinition} "获取成功"
// @Failure 400 {object} APIResponse "工具名称未提供"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 404 {object} APIResponse "工具不存在"
// @Router /api/auth/ai/tools/{name} [get]
func GetTool(c *gin.Context) {
	globalAIHandler.GetTool(c)
}

// HealthCheck AI服务健康检查
// @Summary AI服务健康状态检查
// @Description 检查AI服务和相关组件的健康状态
// @Tags AI工具链
// @Produce json
// @Security BearerAuth
// @Success 200 {object} APIResponse{data=object{status=string}} "服务健康"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 503 {object} APIResponse "服务不可用"
// @Router /api/auth/ai/health [get]
func HealthCheck(c *gin.Context) {
	globalAIHandler.HealthCheck(c)
}

// AIHandler handles AI-related HTTP requests
type AIHandler struct {
	aiService *AIService
}

// NewAIHandler creates a new AI handler
func NewAIHandler() *AIHandler {
	return &AIHandler{
		aiService: NewAIService(),
	}
}

// QueryRequest represents a user query request
type QueryRequest struct {
	Query string `json:"query" binding:"required" example:"连接到数据库 ./data.db 并显示所有表"` // 自然语言查询内容
}

// ToolCallRequest represents a direct tool call request
type ToolCallRequest struct {
	ToolCalls string `json:"tool_calls" binding:"required" example:"[{\"id\":\"call_1\",\"type\":\"function\",\"function\":{\"name\":\"connect_sqlite_database\",\"arguments\":{\"database_path\":\"./test.db\"}}}]"` // JSON格式的工具调用数组
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success" example:"true"`     // 操作是否成功
	Data    interface{} `json:"data,omitempty"`             // 响应数据
	Error   string      `json:"error,omitempty" example:""` // 错误信息
}

// ProcessQuery handles user queries with AI
func (h *AIHandler) ProcessQuery(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	response, err := h.aiService.ProcessUserQuery(req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to process query: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"response": response,
		},
	})
}

// ExecuteToolCalls handles direct tool call execution
func (h *AIHandler) ExecuteToolCalls(c *gin.Context) {
	var req ToolCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	result, err := h.aiService.ProcessWithDirectToolCall(req.ToolCalls)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to execute tool calls: " + err.Error(),
		})
		return
	}

	// Parse result as JSON for structured response
	var structuredResult interface{}
	if err := json.Unmarshal([]byte(result), &structuredResult); err != nil {
		// If parsing fails, return as string
		structuredResult = result
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    structuredResult,
	})
}

// GetAvailableTools returns the list of available tools
func (h *AIHandler) GetAvailableTools(c *gin.Context) {
	tools := h.aiService.GetAvailableTools()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"tools": tools,
			"count": len(tools),
		},
	})
}

// GetTool returns a specific tool by name
func (h *AIHandler) GetTool(c *gin.Context) {
	toolName := c.Param("name")
	if toolName == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Tool name is required",
		})
		return
	}

	tool, err := h.aiService.GetToolByName(toolName)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    tool,
	})
}

// HealthCheck checks the health of AI service
func (h *AIHandler) HealthCheck(c *gin.Context) {
	err := h.aiService.HealthCheck()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, APIResponse{
			Success: false,
			Error:   "AI service is not healthy: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status": "healthy",
		},
	})
}

// RegisterRoutes registers AI-related routes
func (h *AIHandler) RegisterRoutes(router *gin.Engine) {
	aiGroup := router.Group("/api/ai")
	{
		aiGroup.POST("/query", h.ProcessQuery)
		aiGroup.POST("/execute", h.ExecuteToolCalls)
		aiGroup.GET("/tools", h.GetAvailableTools)
		aiGroup.GET("/tools/:name", h.GetTool)
		aiGroup.GET("/health", h.HealthCheck)
	}
}
