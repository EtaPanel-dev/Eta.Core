package pty

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/extend/ws"
	"github.com/gorilla/websocket"
)

type PTY struct {
	ID      string
	Cmd     *exec.Cmd
	Stdin   io.WriteCloser
	Stdout  io.ReadCloser
	Stderr  io.ReadCloser
	Running bool
	mutex   sync.Mutex
	conn    *ws.Connection
}

type PTYManager struct {
	ptys  map[string]*PTY
	mutex sync.RWMutex
}

type CommandRequest struct {
	Command string `json:"command"`
	Args    string `json:"args"`
	Dir     string `json:"dir"`
}

type OutputMessage struct {
	Type   string `json:"type"` // stdout, stderr, exit
	Data   string `json:"data"`
	Code   int    `json:"code"`   // exit code
	PID    int    `json:"pid"`    // process id
	Status string `json:"status"` // running, finished, error
}

// PTYHandler PTY的WebSocket处理器
type PTYHandler struct {
	manager *PTYManager
}

var (
	ptyManager = &PTYManager{
		ptys: make(map[string]*PTY),
	}
)

// GetPTYManager 获取PTY管理器实例
func GetPTYManager() *PTYManager {
	return ptyManager
}

// NewPTYHandler 创建新的PTY处理器
func NewPTYHandler() *PTYHandler {
	return &PTYHandler{
		manager: ptyManager,
	}
}

// NewPTY 创建新的PTY实例（用于测试）
func NewPTY(sessionID string, conn *ws.Connection) *PTY {
	return &PTY{
		ID:      sessionID,
		Running: false,
		conn:    conn,
	}
}

// HandleConnection 实现ws.ConnectionHandler接口
func (h *PTYHandler) HandleConnection(conn *ws.Connection) error {
	// 创建新的PTY实例
	pty := &PTY{
		ID:      conn.ID,
		Running: false,
		conn:    conn,
	}

	// 添加到管理器
	h.manager.AddPTY(conn.ID, pty)

	// 发送欢迎消息
	welcomeMsg := ws.Message{
		Type: "info",
		Data: OutputMessage{
			Type:   "info",
			Data:   "PTY session started. Session Id: " + conn.ID,
			Status: "connected",
		},
	}

	return conn.SendMessage(welcomeMsg)
}

// HandleMessage 实现ws.ConnectionHandler接口
func (h *PTYHandler) HandleMessage(conn *ws.Connection, messageType int, data []byte) error {
	if messageType != websocket.TextMessage {
		return nil
	}

	var msg ws.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return h.sendError(conn, "Invalid message format")
	}

	pty, exists := h.manager.GetPTY(conn.ID)
	if !exists {
		return h.sendError(conn, "PTY session not found")
	}

	switch msg.Type {
	case "command":
		return h.handleCommand(pty, msg.Data)
	case "input":
		return h.handleInput(pty, msg.Data)
	case "resize":
		return h.handleResize(pty, msg.Data)
	case "ping":
		return h.handlePing(conn)
	default:
		return h.sendError(conn, "Unknown message type: "+msg.Type)
	}
}

// HandleClose 实现ws.ConnectionHandler接口
func (h *PTYHandler) HandleClose(conn *ws.Connection) error {
	if pty, exists := h.manager.GetPTY(conn.ID); exists {
		_ = pty.Stop() // 忽略停止错误
		h.manager.RemovePTY(conn.ID)
	}
	return nil
}

// handleCommand 处理命令执行
func (h *PTYHandler) handleCommand(pty *PTY, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return h.sendError(pty.conn, "Invalid command data format")
	}

	var cmdReq CommandRequest
	if err := json.Unmarshal(dataBytes, &cmdReq); err != nil {
		return h.sendError(pty.conn, "Invalid command request format")
	}

	// 停止当前运行的命令
	if pty.Running {
		_ = pty.Stop() // 忽略停止错误
		time.Sleep(100 * time.Millisecond)
	}

	// 解析参数
	var args []string
	if cmdReq.Args != "" {
		args = parseArgs(cmdReq.Args)
	}

	// 启动新命令
	if err := pty.Start(cmdReq.Command, args, cmdReq.Dir); err != nil {
		return h.sendError(pty.conn, fmt.Sprintf("Failed to start command: %v", err))
	}

	// 开始读取输出
	go pty.ReadOutput()

	// 发送启动成功消息
	startMsg := OutputMessage{
		Type:   "info",
		Data:   fmt.Sprintf("Command started: %s %s", cmdReq.Command, cmdReq.Args),
		Status: "running",
	}
	if pty.Cmd != nil && pty.Cmd.Process != nil {
		startMsg.PID = pty.Cmd.Process.Pid
	}

	return pty.sendMessage(startMsg)
}

// handleInput 处理输入
func (h *PTYHandler) handleInput(pty *PTY, data interface{}) error {
	input, ok := data.(string)
	if !ok {
		return h.sendError(pty.conn, "Invalid input data format")
	}

	if err := pty.WriteInput(input); err != nil {
		return h.sendError(pty.conn, fmt.Sprintf("Failed to write input: %v", err))
	}
	return nil
}

// handleResize 处理终端大小调整
func (h *PTYHandler) handleResize(pty *PTY, data interface{}) error {
	_ = data // 忽略data参数，避免编译警告
	resizeMsg := OutputMessage{
		Type:   "info",
		Data:   "Terminal resize acknowledged",
		Status: "running",
	}
	return pty.sendMessage(resizeMsg)
}

// handlePing 处理ping消息
func (h *PTYHandler) handlePing(conn *ws.Connection) error {
	pongMsg := ws.Message{
		Type: "pong",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
	}
	return conn.SendMessage(pongMsg)
}

// sendError 发送错误消息
func (h *PTYHandler) sendError(conn *ws.Connection, message string) error {
	errorMsg := ws.Message{
		Type: "error",
		Data: OutputMessage{
			Type:   "error",
			Data:   message,
			Status: "error",
		},
	}
	return conn.SendMessage(errorMsg)
}

// PTYManager 方法

// AddPTY 添加PTY实例
func (m *PTYManager) AddPTY(id string, pty *PTY) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ptys[id] = pty
}

// GetPTY 获取PTY实例
func (m *PTYManager) GetPTY(id string) (*PTY, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	pty, exists := m.ptys[id]
	return pty, exists
}

// RemovePTY 移除PTY实例
func (m *PTYManager) RemovePTY(id string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.ptys, id)
}

// GetAllPTYs 获取所有PTY实例
func (m *PTYManager) GetAllPTYs() map[string]*PTY {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make(map[string]*PTY)
	for k, v := range m.ptys {
		result[k] = v
	}
	return result
}

// PTY 实例方法

// Start 启动命令
func (p *PTY) Start(command string, args []string, dir string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.Running {
		return fmt.Errorf("PTY is already running")
	}

	// 创建命令
	p.Cmd = exec.Command(command, args...)
	if dir != "" {
		p.Cmd.Dir = dir
	}

	// 设置环境变量
	p.Cmd.Env = os.Environ()

	// 获取标准输入输出
	var err error
	p.Stdin, err = p.Cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	p.Stdout, err = p.Cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	p.Stderr, err = p.Cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// 启动命令
	if err := p.Cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	p.Running = true
	return nil
}

// Stop 停止命令
func (p *PTY) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.Running || p.Cmd == nil {
		return nil
	}

	// 尝试优雅关闭
	if p.Cmd.Process != nil {
		if runtime.GOOS == "windows" {
			_ = p.Cmd.Process.Kill() // 忽略杀死进程错误
		} else {
			_ = p.Cmd.Process.Signal(syscall.SIGTERM) // 忽略信号发送错误

			// 等待一段时间后强制关闭
			done := make(chan error, 1)
			go func() {
				done <- p.Cmd.Wait()
			}()

			select {
			case <-done:
				// 进程已经结束
			case <-time.After(3 * time.Second):
				// 超时，强制杀死进程
				_ = p.Cmd.Process.Kill() // 忽略杀死进程错误
				<-done
			}
		}
	}

	// 关闭管道
	if p.Stdin != nil {
		_ = p.Stdin.Close()
	}
	if p.Stdout != nil {
		_ = p.Stdout.Close()
	}
	if p.Stderr != nil {
		_ = p.Stderr.Close()
	}

	p.Running = false
	return nil
}

// WriteInput 写入输入
func (p *PTY) WriteInput(input string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.Running || p.Stdin == nil {
		return fmt.Errorf("PTY is not running")
	}

	_, err := p.Stdin.Write([]byte(input))
	return err
}

// ReadOutput 读取输出
func (p *PTY) ReadOutput() {
	// 读取标准输出
	go func() {
		scanner := bufio.NewScanner(p.Stdout)
		for scanner.Scan() {
			if !p.Running {
				break
			}
			msg := OutputMessage{
				Type:   "stdout",
				Data:   scanner.Text(),
				Status: "running",
			}
			if p.Cmd != nil && p.Cmd.Process != nil {
				msg.PID = p.Cmd.Process.Pid
			}
			_ = p.sendMessage(msg) // 忽略发送错误
		}
	}()

	// 读取标准错误
	go func() {
		scanner := bufio.NewScanner(p.Stderr)
		for scanner.Scan() {
			if !p.Running {
				break
			}
			msg := OutputMessage{
				Type:   "stderr",
				Data:   scanner.Text(),
				Status: "running",
			}
			if p.Cmd != nil && p.Cmd.Process != nil {
				msg.PID = p.Cmd.Process.Pid
			}
			_ = p.sendMessage(msg) // 忽略发送错误
		}
	}()

	// 等待命令结束
	go func() {
		if p.Cmd != nil {
			err := p.Cmd.Wait()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				}
			}

			msg := OutputMessage{
				Type:   "exit",
				Data:   fmt.Sprintf("Process exited with code %d", exitCode),
				Code:   exitCode,
				Status: "finished",
			}
			if p.Cmd.Process != nil {
				msg.PID = p.Cmd.Process.Pid
			}
			_ = p.sendMessage(msg) // 忽略发送错误
			p.Running = false
		}
	}()
}

// sendMessage 发送消息
func (p *PTY) sendMessage(msg OutputMessage) error {
	if p.conn == nil {
		return fmt.Errorf("no connection available")
	}

	wsMsg := ws.Message{
		Type: "output",
		Data: msg,
	}

	return p.conn.SendMessage(wsMsg)
}

// parseArgs 解析命令参数
func parseArgs(args string) []string {
	args = strings.TrimSpace(args)
	if args == "" {
		return nil
	}

	var result []string
	var current strings.Builder
	inQuotes := false
	escapeNext := false

	for _, r := range args {
		if escapeNext {
			current.WriteRune(r)
			escapeNext = false
			continue
		}

		switch r {
		case '\\':
			escapeNext = true
		case '"':
			inQuotes = !inQuotes
		case ' ':
			if inQuotes {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
