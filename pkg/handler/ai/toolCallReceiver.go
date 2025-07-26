package ai

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/extend/dbmgr"
)

// ToolCall represents a tool call from AI
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents the function call details
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallResult represents the result of a tool call
type ToolCallResult struct {
	ToolCallID string      `json:"tool_call_id"`
	Success    bool        `json:"success"`
	Result     interface{} `json:"result"`
	Error      string      `json:"error,omitempty"`
}

// ToolCallReceiver handles incoming tool calls from AI
type ToolCallReceiver struct {
	sqliteManager *dbmgr.SQLiteManager
}

// NewToolCallReceiver creates a new tool call receiver
func NewToolCallReceiver() *ToolCallReceiver {
	return &ToolCallReceiver{}
}

// ProcessToolCall processes a single tool call
func (tcr *ToolCallReceiver) ProcessToolCall(toolCall ToolCall) ToolCallResult {
	result := ToolCallResult{
		ToolCallID: toolCall.ID,
		Success:    false,
	}

	switch toolCall.Function.Name {
	case "connect_sqlite_database":
		result = tcr.handleConnectDatabase(toolCall)
	case "get_database_info":
		result = tcr.handleGetDatabaseInfo(toolCall)
	case "list_tables":
		result = tcr.handleListTables(toolCall)
	case "get_table_schema":
		result = tcr.handleGetTableSchema(toolCall)
	case "execute_query":
		result = tcr.handleExecuteQuery(toolCall)
	case "execute_statement":
		result = tcr.handleExecuteStatement(toolCall)
	case "create_table":
		result = tcr.handleCreateTable(toolCall)
	case "drop_table":
		result = tcr.handleDropTable(toolCall)
	case "backup_database":
		result = tcr.handleBackupDatabase(toolCall)
	case "vacuum_database":
		result = tcr.handleVacuumDatabase(toolCall)
	default:
		result.Error = fmt.Sprintf("unknown tool function: %s", toolCall.Function.Name)
	}

	return result
}

// ProcessToolCalls processes multiple tool calls
func (tcr *ToolCallReceiver) ProcessToolCalls(toolCalls []ToolCall) []ToolCallResult {
	var results []ToolCallResult

	for _, toolCall := range toolCalls {
		result := tcr.ProcessToolCall(toolCall)
		results = append(results, result)

		// Log the result
		if result.Success {
			log.Printf("Tool call %s executed successfully", toolCall.Function.Name)
		} else {
			log.Printf("Tool call %s failed: %s", toolCall.Function.Name, result.Error)
		}
	}

	return results
}

// handleConnectDatabase handles database connection
func (tcr *ToolCallReceiver) handleConnectDatabase(toolCall ToolCall) ToolCallResult {
	result := ToolCallResult{ToolCallID: toolCall.ID}

	databasePath, ok := toolCall.Function.Arguments["database_path"].(string)
	if !ok {
		result.Error = "database_path parameter is required and must be a string"
		return result
	}

	config := dbmgr.SQLiteConfig{
		DatabasePath: databasePath,
	}

	tcr.sqliteManager = dbmgr.NewSQLiteManager(config)
	err := tcr.sqliteManager.Connect()
	if err != nil {
		result.Error = fmt.Sprintf("failed to connect to database: %v", err)
		return result
	}

	result.Success = true
	result.Result = map[string]interface{}{
		"message": fmt.Sprintf("Successfully connected to SQLite database at %s", databasePath),
		"path":    databasePath,
	}

	return result
}

// handleGetDatabaseInfo handles getting database information
func (tcr *ToolCallReceiver) handleGetDatabaseInfo(toolCall ToolCall) ToolCallResult {
	result := ToolCallResult{ToolCallID: toolCall.ID}

	if tcr.sqliteManager == nil {
		result.Error = "no active database connection"
		return result
	}

	info, err := tcr.sqliteManager.GetDatabaseInfo()
	if err != nil {
		result.Error = fmt.Sprintf("failed to get database info: %v", err)
		return result
	}

	result.Success = true
	result.Result = info

	return result
}

// handleListTables handles listing tables
func (tcr *ToolCallReceiver) handleListTables(toolCall ToolCall) ToolCallResult {
	result := ToolCallResult{ToolCallID: toolCall.ID}

	if tcr.sqliteManager == nil {
		result.Error = "no active database connection"
		return result
	}

	tables, err := tcr.sqliteManager.GetTables()
	if err != nil {
		result.Error = fmt.Sprintf("failed to list tables: %v", err)
		return result
	}

	result.Success = true
	result.Result = tables

	return result
}

// handleGetTableSchema handles getting table schema
func (tcr *ToolCallReceiver) handleGetTableSchema(toolCall ToolCall) ToolCallResult {
	result := ToolCallResult{ToolCallID: toolCall.ID}

	if tcr.sqliteManager == nil {
		result.Error = "no active database connection"
		return result
	}

	tableName, ok := toolCall.Function.Arguments["table_name"].(string)
	if !ok {
		result.Error = "table_name parameter is required and must be a string"
		return result
	}

	schema, err := tcr.sqliteManager.GetTableSchema(tableName)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get table schema: %v", err)
		return result
	}

	result.Success = true
	result.Result = schema

	return result
}

// handleExecuteQuery handles executing SQL queries
func (tcr *ToolCallReceiver) handleExecuteQuery(toolCall ToolCall) ToolCallResult {
	result := ToolCallResult{ToolCallID: toolCall.ID}

	if tcr.sqliteManager == nil {
		result.Error = "no active database connection"
		return result
	}

	query, ok := toolCall.Function.Arguments["query"].(string)
	if !ok {
		result.Error = "query parameter is required and must be a string"
		return result
	}

	queryResult, err := tcr.sqliteManager.ExecuteQuery(query)
	if err != nil {
		result.Error = fmt.Sprintf("failed to execute query: %v", err)
		return result
	}

	result.Success = true
	result.Result = map[string]interface{}{
		"rows":  queryResult,
		"count": len(queryResult),
	}

	return result
}

// handleExecuteStatement handles executing SQL statements
func (tcr *ToolCallReceiver) handleExecuteStatement(toolCall ToolCall) ToolCallResult {
	result := ToolCallResult{ToolCallID: toolCall.ID}

	if tcr.sqliteManager == nil {
		result.Error = "no active database connection"
		return result
	}

	statement, ok := toolCall.Function.Arguments["statement"].(string)
	if !ok {
		result.Error = "statement parameter is required and must be a string"
		return result
	}

	rowsAffected, err := tcr.sqliteManager.ExecuteStatement(statement)
	if err != nil {
		result.Error = fmt.Sprintf("failed to execute statement: %v", err)
		return result
	}

	result.Success = true
	result.Result = map[string]interface{}{
		"rows_affected": rowsAffected,
		"message":       fmt.Sprintf("Statement executed successfully, %d rows affected", rowsAffected),
	}

	return result
}

// handleCreateTable handles creating tables
func (tcr *ToolCallReceiver) handleCreateTable(toolCall ToolCall) ToolCallResult {
	result := ToolCallResult{ToolCallID: toolCall.ID}

	if tcr.sqliteManager == nil {
		result.Error = "no active database connection"
		return result
	}

	createSQL, ok := toolCall.Function.Arguments["create_sql"].(string)
	if !ok {
		result.Error = "create_sql parameter is required and must be a string"
		return result
	}

	err := tcr.sqliteManager.CreateTable(createSQL)
	if err != nil {
		result.Error = fmt.Sprintf("failed to create table: %v", err)
		return result
	}

	result.Success = true
	result.Result = map[string]interface{}{
		"message": "Table created successfully",
	}

	return result
}

// handleDropTable handles dropping tables
func (tcr *ToolCallReceiver) handleDropTable(toolCall ToolCall) ToolCallResult {
	result := ToolCallResult{ToolCallID: toolCall.ID}

	if tcr.sqliteManager == nil {
		result.Error = "no active database connection"
		return result
	}

	tableName, ok := toolCall.Function.Arguments["table_name"].(string)
	if !ok {
		result.Error = "table_name parameter is required and must be a string"
		return result
	}

	err := tcr.sqliteManager.DropTable(tableName)
	if err != nil {
		result.Error = fmt.Sprintf("failed to drop table: %v", err)
		return result
	}

	result.Success = true
	result.Result = map[string]interface{}{
		"message": fmt.Sprintf("Table '%s' dropped successfully", tableName),
	}

	return result
}

// handleBackupDatabase handles database backup
func (tcr *ToolCallReceiver) handleBackupDatabase(toolCall ToolCall) ToolCallResult {
	result := ToolCallResult{ToolCallID: toolCall.ID}

	if tcr.sqliteManager == nil {
		result.Error = "no active database connection"
		return result
	}

	backupPath, ok := toolCall.Function.Arguments["backup_path"].(string)
	if !ok {
		result.Error = "backup_path parameter is required and must be a string"
		return result
	}

	err := tcr.sqliteManager.BackupDatabase(backupPath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to backup database: %v", err)
		return result
	}

	result.Success = true
	result.Result = map[string]interface{}{
		"message":     "Database backed up successfully",
		"backup_path": backupPath,
	}

	return result
}

// handleVacuumDatabase handles database vacuum
func (tcr *ToolCallReceiver) handleVacuumDatabase(toolCall ToolCall) ToolCallResult {
	result := ToolCallResult{ToolCallID: toolCall.ID}

	if tcr.sqliteManager == nil {
		result.Error = "no active database connection"
		return result
	}

	err := tcr.sqliteManager.VacuumDatabase()
	if err != nil {
		result.Error = fmt.Sprintf("failed to vacuum database: %v", err)
		return result
	}

	result.Success = true
	result.Result = map[string]interface{}{
		"message": "Database vacuumed successfully",
	}

	return result
}

// ParseToolCallsFromJSON parses tool calls from JSON string
func ParseToolCallsFromJSON(jsonStr string) ([]ToolCall, error) {
	var toolCalls []ToolCall
	err := json.Unmarshal([]byte(jsonStr), &toolCalls)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tool calls from JSON: %w", err)
	}
	return toolCalls, nil
}

// FormatResults formats tool call results as JSON
func FormatResults(results []ToolCallResult) (string, error) {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format results: %w", err)
	}
	return string(jsonData), nil
}
