package ai

import (
	"encoding/json"
	"fmt"
)

// ToolDefinition defines a tool that can be called by AI
type ToolDefinition struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition defines the function details
type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolChain manages available tools for AI
type ToolChain struct {
	Tools []ToolDefinition `json:"tools"`
}

// GenerateToolChain creates a comprehensive tool chain for database operations
func GenerateToolChain() *ToolChain {
	toolChain := &ToolChain{
		Tools: []ToolDefinition{
			{
				Type: "function",
				Function: FunctionDefinition{
					Name:        "connect_sqlite_database",
					Description: "连接到SQLite数据库",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"database_path": map[string]interface{}{
								"type":        "string",
								"description": "SQLite数据库文件路径",
							},
						},
						"required": []string{"database_path"},
					},
				},
			},
			{
				Type: "function",
				Function: FunctionDefinition{
					Name:        "get_database_info",
					Description: "获取数据库基本信息，包括大小、表数量等",
					Parameters: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
			{
				Type: "function",
				Function: FunctionDefinition{
					Name:        "list_tables",
					Description: "列出数据库中的所有表及其信息",
					Parameters: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
			{
				Type: "function",
				Function: FunctionDefinition{
					Name:        "get_table_schema",
					Description: "获取指定表的结构信息",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"table_name": map[string]interface{}{
								"type":        "string",
								"description": "表名",
							},
						},
						"required": []string{"table_name"},
					},
				},
			},
			{
				Type: "function",
				Function: FunctionDefinition{
					Name:        "execute_query",
					Description: "执行SQL查询语句（SELECT）",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"query": map[string]interface{}{
								"type":        "string",
								"description": "要执行的SQL查询语句",
							},
						},
						"required": []string{"query"},
					},
				},
			},
			{
				Type: "function",
				Function: FunctionDefinition{
					Name:        "execute_statement",
					Description: "执行SQL语句（INSERT, UPDATE, DELETE）",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"statement": map[string]interface{}{
								"type":        "string",
								"description": "要执行的SQL语句",
							},
						},
						"required": []string{"statement"},
					},
				},
			},
			{
				Type: "function",
				Function: FunctionDefinition{
					Name:        "create_table",
					Description: "创建新表",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"create_sql": map[string]interface{}{
								"type":        "string",
								"description": "CREATE TABLE SQL语句",
							},
						},
						"required": []string{"create_sql"},
					},
				},
			},
			{
				Type: "function",
				Function: FunctionDefinition{
					Name:        "drop_table",
					Description: "删除表",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"table_name": map[string]interface{}{
								"type":        "string",
								"description": "要删除的表名",
							},
						},
						"required": []string{"table_name"},
					},
				},
			},
			{
				Type: "function",
				Function: FunctionDefinition{
					Name:        "backup_database",
					Description: "备份数据库",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"backup_path": map[string]interface{}{
								"type":        "string",
								"description": "备份文件保存路径",
							},
						},
						"required": []string{"backup_path"},
					},
				},
			},
			{
				Type: "function",
				Function: FunctionDefinition{
					Name:        "vacuum_database",
					Description: "优化数据库，清理空间并重建索引",
					Parameters: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		},
	}

	return toolChain
}

// GetToolByName returns a tool definition by name
func (tc *ToolChain) GetToolByName(name string) (*ToolDefinition, error) {
	for _, tool := range tc.Tools {
		if tool.Function.Name == name {
			return &tool, nil
		}
	}
	return nil, fmt.Errorf("tool '%s' not found", name)
}

// ToJSON converts the tool chain to JSON format
func (tc *ToolChain) ToJSON() (string, error) {
	jsonData, err := json.MarshalIndent(tc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool chain: %w", err)
	}
	return string(jsonData), nil
}

// GeneratePromptWithTools creates a prompt that includes available tools
func GeneratePromptWithTools(userQuery string) (string, []ToolDefinition, error) {
	toolChain := GenerateToolChain()

	prompt := fmt.Sprintf(`你是一个数据库管理助手。用户的请求是：%s

你可以使用以下工具来帮助用户：

`, userQuery)

	for _, tool := range toolChain.Tools {
		prompt += fmt.Sprintf("- %s: %s\n", tool.Function.Name, tool.Function.Description)
	}

	prompt += `
请根据用户的需求，选择合适的工具并提供参数。如果需要多个步骤，请按顺序说明需要调用哪些工具。
请以JSON格式返回你的工具调用计划。`

	return prompt, toolChain.Tools, nil
}
