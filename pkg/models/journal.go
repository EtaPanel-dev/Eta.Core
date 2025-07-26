package models

import "time"

// LogQueryRequest 日志查询请求
type LogQueryRequest struct {
	RequestID string `json:"request_id" binding:"required"`
}

// LogQueryResponse 日志查询响应
type LogQueryResponse struct {
	RequestID  string                 `json:"request_id"`
	IPFSHash   string                 `json:"ipfs_hash"`
	TxHash     string                 `json:"tx_hash"`
	LogData    map[string]interface{} `json:"log_data"`
	Verified   bool                   `json:"verified"`
	Timestamp  time.Time              `json:"timestamp"`
	VerifyTime time.Time              `json:"verify_time"`
}

// LogVerificationResponse 日志验证响应
type LogVerificationResponse struct {
	RequestID     string    `json:"request_id"`
	IPFSHash      string    `json:"ipfs_hash"`
	TxHash        string    `json:"tx_hash"`
	IPFSVerified  bool      `json:"ipfs_verified"`
	ChainVerified bool      `json:"chain_verified"`
	DataIntegrity bool      `json:"data_integrity"`
	VerifyTime    time.Time `json:"verify_time"`
	Message       string    `json:"message"`
}

// APILogEntry API请求日志结构
type APILogEntry struct {
	Timestamp    time.Time         `json:"timestamp"`
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	Query        string            `json:"query"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	ClientIP     string            `json:"client_ip"`
	UserAgent    string            `json:"user_agent"`
	ResponseCode int               `json:"response_code"`
	ResponseTime int64             `json:"response_time_ms"`
	RequestID    string            `json:"request_id"`
}

// LogHashEntry 日志哈希记录
type LogHashEntry struct {
	RequestID string    `json:"request_id"`
	IPFSHash  string    `json:"ipfs_hash"`
	TxHash    string    `json:"tx_hash"`
	Timestamp time.Time `json:"timestamp"`
	Verified  bool      `json:"verified"`
}
