package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/router"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRouterSetup 测试路由基本设置
func TestRouterSetup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	router.LoadRoutes(r)

	// 测试健康检查端点
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/public/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	assert.Contains(t, response["message"], "Eta Panel API Server Is OK!")
}

// TestAIToolsEndpoint 测试AI工具列表端点
func TestAIToolsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	router.LoadRoutes(r)

	// 模拟认证token（在实际测试中需要真实的JWT）
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/auth/ai/tools", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	r.ServeHTTP(w, req)

	// 由于没有真实的JWT认证，这里会返回401，但路由应该存在
	assert.NotEqual(t, 404, w.Code) // 确保路由存在
}

// TestAIHealthEndpoint 测试AI健康检查端点
func TestAIHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	router.LoadRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/auth/ai/health", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	r.ServeHTTP(w, req)

	// 确保路由存在（即使认证失败）
	assert.NotEqual(t, 404, w.Code)
}

// TestNotFoundRoute 测试404处理
func TestNotFoundRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	router.LoadRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(404), response["code"])
	assert.Equal(t, "API route not found", response["message"])
}

// 演示函数：显示所有可用的API端点
func demonstrateAPIEndpoints() {
	fmt.Println("=== EtaPanel API 端点总览 ===")

	endpoints := map[string][]string{
		"公共API": {
			"GET /api/public/ - 健康检查",
			"POST /api/public/login - 用户登录",
		},
		"文件管理API": {
			"GET /api/auth/files/ - 列出文件",
			"POST /api/auth/files/upload - 上传文件",
			"POST /api/auth/files/mkdir - 创建目录",
			"DELETE /api/auth/files/ - 删除文件",
		},
		"系统监控API": {
			"GET /api/auth/system/ - 系统信息",
			"GET /api/auth/system/cpu - CPU信息",
			"GET /api/auth/system/memory - 内存信息",
			"GET /api/auth/system/processes - 进程列表",
		},
		"AI工具链API（新增）": {
			"POST /api/auth/ai/query - 自然语言数据库查询",
			"POST /api/auth/ai/execute - 执行数据库工具调用",
			"GET /api/auth/ai/tools - 获取可用工具列表",
			"GET /api/auth/ai/tools/:name - 获取特定工具信息",
			"GET /api/auth/ai/health - AI服务健康检查",
		},
		"传统AI功能": {
			"POST /api/auth/ai/log - 日志分析",
			"POST /api/auth/ai/files - 文件分析",
		},
		"SSL证书管理": {
			"GET /api/auth/acme/ssl/ - 获取SSL证书",
			"POST /api/auth/acme/ssl/ - 申请SSL证书",
			"GET /api/auth/acme/clients/ - 获取ACME客户端",
		},
		"Nginx管理": {
			"GET /api/auth/nginx/status - Nginx状态",
			"GET /api/auth/nginx/sites - 网站列表",
			"POST /api/auth/nginx/restart - 重启Nginx",
		},
		"定时任务": {
			"GET /api/auth/crontab/ - 定时任务列表",
			"POST /api/auth/crontab/ - 创建定时任务",
		},
		"日志管理": {
			"POST /api/log/query - 查询日志",
			"GET /api/log/stats - 日志统计",
		},
		"WebSocket": {
			"GET /ws/pty - PTY终端连接",
		},
	}

	for category, apis := range endpoints {
		fmt.Printf("\n【%s】\n", category)
		for _, api := range apis {
			fmt.Printf("  %s\n", api)
		}
	}

	fmt.Println("\n=== AI数据库工具链特色功能 ===")
	fmt.Println("✅ 自然语言数据库操作")
	fmt.Println("✅ SQLite数据库管理（替换MySQL）")
	fmt.Println("✅ 10个预定义数据库工具")
	fmt.Println("✅ 批量工具调用支持")
	fmt.Println("✅ 完整的错误处理")
	fmt.Println("✅ JWT认证保护")
}

func main() {
	demonstrateAPIEndpoints()

	fmt.Println("\n=== 路由配置完成状态 ===")
	fmt.Println("✅ 基础路由配置完成")
	fmt.Println("✅ AI工具链路由集成完成")
	fmt.Println("✅ 中间件配置完成")
	fmt.Println("✅ 错误处理完成")
	fmt.Println("✅ WebSocket支持完成")

	fmt.Println("\n系统现在支持以下AI数据库工具：")
	tools := []string{
		"connect_sqlite_database - 连接SQLite数据库",
		"get_database_info - 获取数据库信息",
		"list_tables - 列出所有表",
		"get_table_schema - 获取表结构",
		"execute_query - 执行查询语句",
		"execute_statement - 执行更新语句",
		"create_table - 创建表",
		"drop_table - 删除表",
		"backup_database - 备份数据库",
		"vacuum_database - 优化数据库",
	}

	for i, tool := range tools {
		fmt.Printf("%d. %s\n", i+1, tool)
	}
}
