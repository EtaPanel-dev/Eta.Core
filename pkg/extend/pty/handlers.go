package pty

import (
	"net/http"

	"github.com/LxHTT/Eta-Panel/core/pkg/extend/ws"
)

// RegisterPTYHandler 注册PTY WebSocket处理器
func RegisterPTYHandler(path string) http.Handler {
	// 创建PTY处理器
	ptyHandler := NewPTYHandler()

	// 注册到WebSocket管理器
	return ws.RegisterHandler(path, ptyHandler)
}

// InitPTYService 初始化PTY服务
func InitPTYService() {
	// 可以在这里进行一些初始化工作
	// 比如设置默认配置、清理资源等
}
