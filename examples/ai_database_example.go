package main

import (
	"fmt"
	"log"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/ai"
)

func main() {
	// 示例1: 直接使用工具调用
	fmt.Println("=== 示例1: 直接工具调用 ===")

	// 创建工具调用接收器
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

	// 获取数据库信息
	infoCall := ai.ToolCall{
		ID:   "call_2",
		Type: "function",
		Function: ai.FunctionCall{
			Name:      "get_database_info",
			Arguments: map[string]interface{}{},
		},
	}

	result = receiver.ProcessToolCall(infoCall)
	fmt.Printf("数据库信息: %+v\n", result)

	// 示例2: 使用AI服务
	fmt.Println("\n=== 示例2: AI自然语言查询 ===")

	aiService := ai.NewAIService()

	// 自然语言查询示例
	queries := []string{
		"连接到数据库 ./test.db",
		"显示数据库信息",
		"列出所有表",
		"创建一个用户表，包含id、name、email字段",
	}

	for _, query := range queries {
		fmt.Printf("\n用户查询: %s\n", query)
		response, err := aiService.ProcessUserQuery(query)
		if err != nil {
			log.Printf("处理查询失败: %v", err)
			continue
		}
		fmt.Printf("AI响应: %s\n", response)
	}

	// 示例3: 获取可用工具
	fmt.Println("\n=== 示例3: 可用工具列表 ===")

	tools := aiService.GetAvailableTools()
	fmt.Printf("共有 %d 个可用工具:\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("- %s: %s\n", tool.Function.Name, tool.Function.Description)
	}
}
