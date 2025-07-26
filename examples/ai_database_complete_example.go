package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/ai"
	"github.com/gin-gonic/gin"
)

func main() {
	// 示例1: 直接使用工具调用
	fmt.Println("=== 示例1: 直接工具调用 ===")

	receiver := ai.NewToolCallReceiver()

	// 连接数据库
	connectCall := ai.ToolCall{
		ID:   "call_1",
		Type: "function",
		Function: ai.FunctionCall{
			Name: "connect_sqlite_database",
			Arguments: map[string]interface{}{
				"database_path": "./test.db",
			},
		},
	}

	result := receiver.ProcessToolCall(connectCall)
	fmt.Printf("连接结果: %+v\n", result)

	// 创建用户表
	createTableCall := ai.ToolCall{
		ID:   "call_2",
		Type: "function",
		Function: ai.FunctionCall{
			Name: "create_table",
			Arguments: map[string]interface{}{
				"create_sql": `CREATE TABLE IF NOT EXISTS users (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL,
					email TEXT UNIQUE NOT NULL,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP
				)`,
			},
		},
	}

	result = receiver.ProcessToolCall(createTableCall)
	fmt.Printf("创建表结果: %+v\n", result)

	// 插入测试数据
	insertCall := ai.ToolCall{
		ID:   "call_3",
		Type: "function",
		Function: ai.FunctionCall{
			Name: "execute_statement",
			Arguments: map[string]interface{}{
				"statement": "INSERT INTO users (name, email) VALUES ('张三', 'zhangsan@example.com'), ('李四', 'lisi@example.com')",
			},
		},
	}

	result = receiver.ProcessToolCall(insertCall)
	fmt.Printf("插入数据结果: %+v\n", result)

	// 查询数据
	queryCall := ai.ToolCall{
		ID:   "call_4",
		Type: "function",
		Function: ai.FunctionCall{
			Name: "execute_query",
			Arguments: map[string]interface{}{
				"query": "SELECT * FROM users",
			},
		},
	}

	result = receiver.ProcessToolCall(queryCall)
	fmt.Printf("查询结果: %+v\n", result)

	// 示例2: 批量处理工具调用
	fmt.Println("\n=== 示例2: 批量工具调用 ===")

	batchCalls := []ai.ToolCall{
		{
			ID:   "batch_1",
			Type: "function",
			Function: ai.FunctionCall{
				Name:      "get_database_info",
				Arguments: map[string]interface{}{},
			},
		},
		{
			ID:   "batch_2",
			Type: "function",
			Function: ai.FunctionCall{
				Name:      "list_tables",
				Arguments: map[string]interface{}{},
			},
		},
		{
			ID:   "batch_3",
			Type: "function",
			Function: ai.FunctionCall{
				Name: "get_table_schema",
				Arguments: map[string]interface{}{
					"table_name": "users",
				},
			},
		},
	}

	batchResults := receiver.ProcessToolCalls(batchCalls)
	for i, result := range batchResults {
		fmt.Printf("批量调用 %d 结果: %+v\n", i+1, result)
	}

	// 示例3: 使用AI服务 (模拟)
	fmt.Println("\n=== 示例3: AI工具链生成 ===")

	toolChain := ai.GenerateToolChain()
	toolsJSON, _ := toolChain.ToJSON()
	fmt.Printf("可用工具链:\n%s\n", toolsJSON)

	// 示例4: JSON格式的工具调用
	fmt.Println("\n=== 示例4: JSON格式工具调用 ===")

	jsonToolCalls := `[
		{
			"id": "json_1",
			"type": "function",
			"function": {
				"name": "execute_query",
				"arguments": {
					"query": "SELECT COUNT(*) as user_count FROM users"
				}
			}
		}
	]`

	aiService := ai.NewAIService()
	jsonResult, err := aiService.ProcessWithDirectToolCall(jsonToolCalls)
	if err != nil {
		log.Printf("JSON工具调用失败: %v", err)
	} else {
		fmt.Printf("JSON工具调用结果:\n%s\n", jsonResult)
	}

	// 示例5: 启动HTTP服务器演示API
	fmt.Println("\n=== 示例5: 启动HTTP API服务器 ===")
	startHTTPServer()
}

func startHTTPServer() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 创建AI处理器并注册路由
	aiHandler := ai.NewAIHandler()
	aiHandler.RegisterRoutes(router)

	// 添加一些示例路由
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "EtaPanel AI Database API",
			"version": "1.0.0",
			"endpoints": []string{
				"POST /api/ai/query - 自然语言查询",
				"POST /api/ai/execute - 直接执行工具调用",
				"GET /api/ai/tools - 获取可用工具",
				"GET /api/ai/tools/:name - 获取特定工具信息",
				"GET /api/ai/health - 健康检查",
			},
		})
	})

	// 示例API调用
	router.GET("/examples", func(c *gin.Context) {
		examples := map[string]interface{}{
			"query_example": map[string]interface{}{
				"method": "POST",
				"url":    "/api/ai/query",
				"body": map[string]string{
					"query": "连接到数据库 ./test.db 并显示所有表",
				},
			},
			"execute_example": map[string]interface{}{
				"method": "POST",
				"url":    "/api/ai/execute",
				"body": map[string]string{
					"tool_calls": `[{"id":"1","type":"function","function":{"name":"connect_sqlite_database","arguments":{"database_path":"./test.db"}}}]`,
				},
			},
		}
		c.JSON(http.StatusOK, examples)
	})

	fmt.Printf("HTTP服务器启动在端口 8080\n")
	fmt.Printf("API文档: http://localhost:8080/examples\n")
	fmt.Printf("按 Ctrl+C 停止服务器\n")

	if err := router.Run(":8080"); err != nil {
		log.Printf("启动服务器失败: %v", err)
	}
}
