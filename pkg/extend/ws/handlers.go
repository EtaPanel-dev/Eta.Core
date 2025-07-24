package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源（生产环境中应该更严格）
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HTTPHandler WebSocket HTTP处理器
type HTTPHandler struct {
	manager *Manager
	path    string
}

// NewHTTPHandler 创建新的HTTP处理器
func NewHTTPHandler(path string) *HTTPHandler {
	return &HTTPHandler{
		manager: GetManager(),
		path:    path,
	}
}

// ServeHTTP 实现http.Handler接口
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 升级HTTP连接到WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusBadRequest)
		return
	}

	// 创建WebSocket连接
	_, err = h.manager.CreateConnection(conn, h.path)
	if err != nil {
		conn.Close()
		http.Error(w, "Failed to create WebSocket connection", http.StatusInternalServerError)
		return
	}
}

// RegisterHandler 注册处理器到指定路径
func RegisterHandler(path string, handler ConnectionHandler) *HTTPHandler {
	manager := GetManager()
	manager.RegisterHandler(path, handler)
	return NewHTTPHandler(path)
}
