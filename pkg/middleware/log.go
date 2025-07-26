package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/blockchain"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models"
	"github.com/gin-gonic/gin"
	shell "github.com/ipfs/go-ipfs-api"
)

// IPFS客户端配置
var ipfsClient *shell.Shell

// 初始化IPFS客户端
func InitIPFS(ipfsURL string) {
	if ipfsURL == "" {
		ipfsURL = "localhost:5001" // 默认IPFS节点地址
	}
	ipfsClient = shell.NewShell(ipfsURL)
}

// APILogger API请求日志记录中间件
func APILogger() gin.HandlerFunc {
	// 确保IPFS客户端已初始化
	if ipfsClient == nil {
		InitIPFS("")
	}

	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 异步处理日志上传，避免阻塞请求
		go func() {
			if err := processAPILog(param); err != nil {
				fmt.Printf("处理API日志失败: %v\n", err)
			}
		}()

		// 返回标准日志格式
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s\" %s\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
		)
	})
}

// 处理API日志
func processAPILog(param gin.LogFormatterParams) error {
	// 生成请求ID
	requestID := generateRequestID(param)

	// 读取请求体
	var bodyBytes []byte
	if param.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(param.Request.Body)
		param.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 构建日志条目
	logEntry := models.APILogEntry{
		Timestamp:    param.TimeStamp,
		Method:       param.Method,
		Path:         param.Path,
		Query:        param.Request.URL.RawQuery,
		Headers:      extractHeaders(param.Request.Header),
		Body:         string(bodyBytes),
		ClientIP:     param.ClientIP,
		UserAgent:    param.Request.UserAgent(),
		ResponseCode: param.StatusCode,
		ResponseTime: param.Latency.Milliseconds(),
		RequestID:    requestID,
	}

	// 将日志转换为JSON
	logJSON, err := json.MarshalIndent(logEntry, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化日志失败: %v", err)
	}

	// 上传到IPFS
	ipfsHash, err := uploadToIPFS(logJSON)
	if err != nil {
		return fmt.Errorf("上传到IPFS失败: %v", err)
	}

	// 上传哈希到Injective Web3
	txHash, err := uploadHashToInjective(requestID, ipfsHash)
	if err != nil {
		return fmt.Errorf("上传到Injective失败: %v", err)
	}

	// 记录哈希信息
	hashEntry := models.LogHashEntry{
		RequestID: requestID,
		IPFSHash:  ipfsHash,
		TxHash:    txHash,
		Timestamp: time.Now(),
		Verified:  true,
	}

	// 保存哈希记录到本地（可选）
	if err := saveHashEntry(hashEntry); err != nil {
		fmt.Printf("保存哈希记录失败: %v\n", err)
	}

	fmt.Printf("API日志已处理: RequestID=%s, IPFS=%s, TxHash=%s\n",
		requestID, ipfsHash, txHash)

	return nil
}

// 生成请求ID
func generateRequestID(param gin.LogFormatterParams) string {
	data := fmt.Sprintf("%s-%s-%s-%d",
		param.ClientIP, param.Method, param.Path, param.TimeStamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

// 提取请求头
func extractHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		// 过滤敏感头信息
		if isSensitiveHeader(key) {
			result[key] = "[FILTERED]"
		} else {
			result[key] = strings.Join(values, ", ")
		}
	}
	return result
}

// 检查是否为敏感头信息
func isSensitiveHeader(header string) bool {
	sensitiveHeaders := []string{
		"authorization", "cookie", "x-api-key", "x-auth-token",
		"x-access-token", "x-csrf-token", "x-session-id",
	}

	header = strings.ToLower(header)
	for _, sensitive := range sensitiveHeaders {
		if header == sensitive {
			return true
		}
	}
	return false
}

// 上传到IPFS
func uploadToIPFS(data []byte) (string, error) {
	if ipfsClient == nil {
		return "", fmt.Errorf("IPFS客户端未初始化")
	}

	reader := bytes.NewReader(data)
	hash, err := ipfsClient.Add(reader)
	if err != nil {
		return "", fmt.Errorf("IPFS上传失败: %v", err)
	}

	return hash, nil
}

// 全局Injective客户端
var injectiveClient InjectiveClient

// InitInjectiveClient 初始化Injective客户端
func InitInjectiveClient(cfg *models.InjectiveConfig) {
	injectiveClient = blockchain.NewInjectiveClient(cfg)
}

// InjectiveClient Injective客户端接口
type InjectiveClient interface {
	UploadLogHash(requestID, ipfsHash string) (string, error)
}

// MockInjectiveClient 模拟Injective客户端
type MockInjectiveClient struct{}

// NewInjectiveClient 创建Injective客户端
func NewInjectiveClient(cfg *models.InjectiveConfig) InjectiveClient {
	// 这里可以根据配置返回真实客户端或模拟客户端
	return &MockInjectiveClient{}
}

// UploadLogHash 上传日志哈希（模拟实现）
func (c *MockInjectiveClient) UploadLogHash(requestID, ipfsHash string) (string, error) {
	// 构建交易数据
	txData := map[string]interface{}{
		"request_id": requestID,
		"ipfs_hash":  ipfsHash,
		"timestamp":  time.Now().Unix(),
		"type":       "api_log_hash",
	}

	// 模拟交易哈希生成
	txDataBytes, _ := json.Marshal(txData)
	hash := sha256.Sum256(txDataBytes)
	txHash := hex.EncodeToString(hash[:])

	fmt.Printf("模拟上传到Injective: RequestID=%s, IPFS=%s, TxHash=%s\n",
		requestID, ipfsHash, txHash)

	return txHash, nil
}

// 上传哈希到Injective Web3
func uploadHashToInjective(requestID, ipfsHash string) (string, error) {
	if injectiveClient == nil {
		// 使用默认模拟客户端
		injectiveClient = NewInjectiveClient(&models.InjectiveConfig{Enabled: true})
	}

	return injectiveClient.UploadLogHash(requestID, ipfsHash)
}

// 保存哈希记录到本地文件
func saveHashEntry(entry models.LogHashEntry) error {
	// 调用handlers包中的函数保存记录
	// 注意：这里需要导入handlers包，但为了避免循环依赖，
	// 我们使用一个全局的存储映射
	logHashStore[entry.RequestID] = entry
	return nil
}

// 全局日志哈希存储
var logHashStore = make(map[string]models.LogHashEntry)

// GetLogHashStore 获取日志哈希存储（供handlers使用）
func GetLogHashStore() map[string]models.LogHashEntry {
	return logHashStore
}

// LogVerification 日志验证中间件
func LogVerification() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在响应头中添加日志哈希信息
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID(gin.LogFormatterParams{
				ClientIP:  c.ClientIP(),
				Method:    c.Request.Method,
				Path:      c.Request.URL.Path,
				TimeStamp: time.Now(),
			})
			c.Header("X-Request-ID", requestID)
		}

		c.Next()

		// 在响应中添加日志验证信息
		c.Header("X-Log-Verification", "enabled")
		c.Header("X-Log-Request-ID", requestID)
	}
}
