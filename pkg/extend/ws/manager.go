package ws

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Connection 表示一个WebSocket连接
type Connection struct {
	ID      string
	Conn    *websocket.Conn
	Send    chan []byte
	Handler ConnectionHandler
	manager *Manager
	mutex   sync.Mutex
	closed  bool
}

// Message 表示WebSocket消息
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// ConnectionHandler 定义连接处理器接口
type ConnectionHandler interface {
	HandleConnection(conn *Connection) error
	HandleMessage(conn *Connection, messageType int, data []byte) error
	HandleClose(conn *Connection) error
}

// Manager WebSocket连接管理器
type Manager struct {
	connections map[string]*Connection
	handlers    map[string]ConnectionHandler
	register    chan *Connection
	unregister  chan *Connection
	mutex       sync.RWMutex
	running     bool
}

var (
	globalManager *Manager
	once          sync.Once
)

// GetManager 获取全局WebSocket管理器实例
func GetManager() *Manager {
	once.Do(func() {
		globalManager = NewManager()
		globalManager.Start()
	})
	return globalManager
}

// NewManager 创建新的WebSocket管理器
func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*Connection),
		handlers:    make(map[string]ConnectionHandler),
		register:    make(chan *Connection, 256),
		unregister:  make(chan *Connection, 256),
	}
}

// RegisterHandler 注册连接处理器
func (m *Manager) RegisterHandler(path string, handler ConnectionHandler) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.handlers[path] = handler
}

// CreateConnection 创建新的WebSocket连接
func (m *Manager) CreateConnection(conn *websocket.Conn, path string) (*Connection, error) {
	handler, exists := m.handlers[path]
	if !exists {
		return nil, fmt.Errorf("no handler registered for path: %s", path)
	}

	connection := &Connection{
		ID:      uuid.New().String(),
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Handler: handler,
		manager: m,
	}

	// 注册连接
	m.register <- connection

	return connection, nil
}

// Start 启动管理器
func (m *Manager) Start() {
	if m.running {
		return
	}
	m.running = true
	go m.run()
}

// Stop 停止管理器
func (m *Manager) Stop() {
	m.running = false
	m.mutex.Lock()
	for _, conn := range m.connections {
		conn.close()
	}
	m.mutex.Unlock()
}

// run 运行管理器主循环
func (m *Manager) run() {
	for {
		select {
		case conn := <-m.register:
			m.handleRegister(conn)
		case conn := <-m.unregister:
			m.handleUnregister(conn)
		}
	}
}

// handleRegister 处理连接注册
func (m *Manager) handleRegister(conn *Connection) {
	m.mutex.Lock()
	m.connections[conn.ID] = conn
	m.mutex.Unlock()

	// 启动连接处理
	go conn.readPump()
	go conn.writePump()

	// 调用处理器的连接回调
	if err := conn.Handler.HandleConnection(conn); err != nil {
		fmt.Printf("Handler connection error: %v\n", err)
	}
}

// handleUnregister 处理连接注销
func (m *Manager) handleUnregister(conn *Connection) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.connections[conn.ID]; exists {
		delete(m.connections, conn.ID)
		close(conn.Send)

		// 调用处理器的关闭回调
		if err := conn.Handler.HandleClose(conn); err != nil {
			fmt.Printf("Handler close error: %v\n", err)
		}
	}
}

// GetConnection 根据ID获取连接
func (m *Manager) GetConnection(id string) (*Connection, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	conn, exists := m.connections[id]
	return conn, exists
}

// GetConnectionCount 获取连接数量
func (m *Manager) GetConnectionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.connections)
}

// BroadcastToPath 向指定路径的所有连接广播消息
func (m *Manager) BroadcastToPath(path string, message Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, conn := range m.connections {
		// TODO: 可以根据需要添加路径过滤逻辑
		_ = path // 暂时忽略path参数，避免编译警告
		select {
		case conn.Send <- data:
		default:
			// 如果发送缓冲区满，关闭连接
			go conn.close()
		}
	}

	return nil
}

// SendMessage 发送消息到连接
func (c *Connection) SendMessage(message Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return fmt.Errorf("connection is closed")
	}

	select {
	case c.Send <- data:
		return nil
	default:
		return fmt.Errorf("send buffer is full")
	}
}

// Close 关闭连接
func (c *Connection) Close() {
	c.close()
}

// close 内部关闭方法
func (c *Connection) close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.closed {
		c.closed = true
		_ = c.Conn.Close() // 忽略关闭错误
		c.manager.unregister <- c
	}
}

// readPump 读取消息泵
func (c *Connection) readPump() {
	defer c.close()

	// 设置读取超时
	_ = c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		messageType, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		// 调用处理器的消息回调
		if err := c.Handler.HandleMessage(c, messageType, data); err != nil {
			fmt.Printf("Handler message error: %v\n", err)
		}
	}
}

// writePump 写入消息泵
func (c *Connection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
