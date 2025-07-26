package log

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/middleware"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models"
	"github.com/gin-gonic/gin"
	shell "github.com/ipfs/go-ipfs-api"
)

// GetLogByRequestID 根据请求ID获取日志
func GetLogByRequestID(c *gin.Context) {
	var req models.LogQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	// 查找日志哈希记录
	logHashStore := middleware.GetLogHashStore()
	hashEntry, exists := logHashStore[req.RequestID]
	if !exists {
		handler.Respond(c, http.StatusNotFound, "未找到指定的日志记录", nil)
		return
	}

	// 从IPFS获取日志数据
	logData, err := getLogFromIPFS(hashEntry.IPFSHash)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "从IPFS获取日志失败: "+err.Error(), nil)
		return
	}

	response := models.LogQueryResponse{
		RequestID:  req.RequestID,
		IPFSHash:   hashEntry.IPFSHash,
		TxHash:     hashEntry.TxHash,
		LogData:    logData,
		Verified:   hashEntry.Verified,
		Timestamp:  hashEntry.Timestamp,
		VerifyTime: time.Now(),
	}

	handler.Respond(c, http.StatusOK, "获取日志成功", response)
}

// VerifyLogIntegrity 验证日志完整性
func VerifyLogIntegrity(c *gin.Context) {
	var req models.LogQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	// 查找日志哈希记录
	logHashStore := middleware.GetLogHashStore()
	hashEntry, exists := logHashStore[req.RequestID]
	if !exists {
		handler.Respond(c, http.StatusNotFound, "未找到指定的日志记录", nil)
		return
	}

	// 验证IPFS数据
	ipfsVerified, err := verifyIPFSData(hashEntry.IPFSHash)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "IPFS验证失败: "+err.Error(), nil)
		return
	}

	// 验证链上数据
	chainVerified, err := verifyChainData(hashEntry.TxHash, hashEntry.IPFSHash)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "链上验证失败: "+err.Error(), nil)
		return
	}

	// 验证数据完整性
	dataIntegrity := ipfsVerified && chainVerified

	var message string
	if dataIntegrity {
		message = "日志数据完整性验证通过"
	} else {
		message = "日志数据完整性验证失败"
	}

	response := models.LogVerificationResponse{
		RequestID:     req.RequestID,
		IPFSHash:      hashEntry.IPFSHash,
		TxHash:        hashEntry.TxHash,
		IPFSVerified:  ipfsVerified,
		ChainVerified: chainVerified,
		DataIntegrity: dataIntegrity,
		VerifyTime:    time.Now(),
		Message:       message,
	}

	handler.Respond(c, http.StatusOK, "验证完成", response)
}

// ListLogHashes 获取所有日志哈希列表
func ListLogHashes(c *gin.Context) {
	logHashStore := middleware.GetLogHashStore()
	var hashes []models.LogHashEntry
	for _, entry := range logHashStore {
		hashes = append(hashes, entry)
	}

	handler.Respond(c, http.StatusOK, "获取日志哈希列表成功", hashes)
}

// GetLogStats 获取日志统计信息
func GetLogStats(c *gin.Context) {
	logHashStore := middleware.GetLogHashStore()
	stats := map[string]interface{}{
		"total_logs":    len(logHashStore),
		"verified_logs": countVerifiedLogs(logHashStore),
		"ipfs_enabled":  true,
		"last_updated":  time.Now(),
	}

	handler.Respond(c, http.StatusOK, "获取日志统计成功", stats)
}

// 从IPFS获取日志数据
func getLogFromIPFS(hash string) (map[string]interface{}, error) {
	ipfsClient := shell.NewShell("localhost:5001")

	reader, err := ipfsClient.Cat(hash)
	if err != nil {
		return nil, fmt.Errorf("从IPFS读取数据失败: %v", err)
	}
	defer reader.Close()

	var logData map[string]interface{}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&logData); err != nil {
		return nil, fmt.Errorf("解析日志数据失败: %v", err)
	}

	return logData, nil
}

// 验证IPFS数据
func verifyIPFSData(hash string) (bool, error) {
	ipfsClient := shell.NewShell("localhost:5001")

	reader, err := ipfsClient.Cat(hash)
	if err != nil {
		return false, nil
	}
	defer reader.Close()

	return true, nil
}

// 验证链上数据（模拟实现）
func verifyChainData(txHash, ipfsHash string) (bool, error) {
	fmt.Printf("模拟验证链上数据: TxHash=%s, IPFSHash=%s\n", txHash, ipfsHash)
	return len(txHash) > 0 && len(ipfsHash) > 0, nil
}

// 统计已验证的日志数量
func countVerifiedLogs(logHashStore map[string]models.LogHashEntry) int {
	count := 0
	for _, entry := range logHashStore {
		if entry.Verified {
			count++
		}
	}
	return count
}
