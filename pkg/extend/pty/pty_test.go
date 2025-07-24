package pty

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// MockConnection 模拟连接，用于不使用WebSocket的测试
type MockConnection struct {
	ID       string
	messages []OutputMessage
	mutex    sync.Mutex
	closed   bool
}

func NewMockConnection(id string) *MockConnection {
	return &MockConnection{
		ID:       id,
		messages: make([]OutputMessage, 0),
		closed:   false,
	}
}

func (mc *MockConnection) SendMessage(data interface{}) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if mc.closed {
		return fmt.Errorf("connection is closed")
	}

	if msg, ok := data.(OutputMessage); ok {
		mc.messages = append(mc.messages, msg)
	}
	return nil
}

func (mc *MockConnection) GetMessages() []OutputMessage {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	result := make([]OutputMessage, len(mc.messages))
	copy(result, mc.messages)
	return result
}

func (mc *MockConnection) ClearMessages() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.messages = mc.messages[:0]
}

func (mc *MockConnection) Close() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.closed = true
}

// StandalonePTY 独立的PTY实现，不依赖WebSocket
type StandalonePTY struct {
	*PTY
	mockConn *MockConnection
}

// NewStandalonePTY 创建独立的PTY实例
func NewStandalonePTY(sessionID string) *StandalonePTY {
	mockConn := NewMockConnection(sessionID)
	pty := &PTY{
		ID:      sessionID,
		Running: false,
		conn:    nil, // 不使用WebSocket连接
	}

	standalonePTY := &StandalonePTY{
		PTY:      pty,
		mockConn: mockConn,
	}

	return standalonePTY
}

// sendMessage 重写发送消息方法
func (sp *StandalonePTY) sendMessage(msg OutputMessage) error {
	return sp.mockConn.SendMessage(msg)
}

// GetMessages 获取所有消息
func (sp *StandalonePTY) GetMessages() []OutputMessage {
	return sp.mockConn.GetMessages()
}

// ClearMessages 清空消息
func (sp *StandalonePTY) ClearMessages() {
	sp.mockConn.ClearMessages()
}

// ReadOutputToMock 读取输出到模拟连接
func (sp *StandalonePTY) ReadOutputToMock() {
	// 读取标准输出
	go func() {
		if sp.Stdout == nil {
			return
		}
		scanner := bufio.NewScanner(sp.Stdout)
		for scanner.Scan() {
			if !sp.Running {
				break
			}
			msg := OutputMessage{
				Type:   "stdout",
				Data:   scanner.Text(),
				Status: "running",
			}
			if sp.Cmd != nil && sp.Cmd.Process != nil {
				msg.PID = sp.Cmd.Process.Pid
			}
			_ = sp.sendMessage(msg)
		}
	}()

	// 读取标准错误
	go func() {
		if sp.Stderr == nil {
			return
		}
		scanner := bufio.NewScanner(sp.Stderr)
		for scanner.Scan() {
			if !sp.Running {
				break
			}
			msg := OutputMessage{
				Type:   "stderr",
				Data:   scanner.Text(),
				Status: "running",
			}
			if sp.Cmd != nil && sp.Cmd.Process != nil {
				msg.PID = sp.Cmd.Process.Pid
			}
			_ = sp.sendMessage(msg)
		}
	}()

	// 等待命令结束
	go func() {
		if sp.Cmd != nil {
			err := sp.Cmd.Wait()
			exitCode := 0
			if err != nil {
				var exitError *exec.ExitError
				if errors.As(err, &exitError) {
					exitCode = exitError.ExitCode()
				}
			}

			msg := OutputMessage{
				Type:   "exit",
				Data:   fmt.Sprintf("Process exited with code %d", exitCode),
				Code:   exitCode,
				Status: "finished",
			}
			if sp.Cmd.Process != nil {
				msg.PID = sp.Cmd.Process.Pid
			}
			_ = sp.sendMessage(msg)
			sp.Running = false
		}
	}()
}

// 测试函数

func TestStandalonePTY_BasicCommand(t *testing.T) {
	pty := NewStandalonePTY("test-session-1")
	defer func(pty *StandalonePTY) {
		err := pty.Stop()
		if err != nil {

		}
	}(pty)

	// 根据操作系统选择测试命令
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "Hello World"}
	} else {
		cmd = "echo"
		args = []string{"Hello World"}
	}

	// 启动命令
	err := pty.Start(cmd, args, "")
	if err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	// 开始读取输出
	pty.ReadOutputToMock()

	// 等待命令完成
	time.Sleep(2 * time.Second)

	// 检查消息
	messages := pty.GetMessages()
	if len(messages) == 0 {
		t.Fatal("No messages received")
	}

	// 查找stdout消息
	found := false
	for _, msg := range messages {
		if msg.Type == "stdout" && strings.Contains(msg.Data, "Hello World") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected stdout message with 'Hello World' not found")
		for i, msg := range messages {
			t.Logf("Message %d: Type=%s, Data=%s, Status=%s", i, msg.Type, msg.Data, msg.Status)
		}
	}
}

func TestStandalonePTY_MultipleCommands(t *testing.T) {
	pty := NewStandalonePTY("test-session-2")
	defer func(pty *StandalonePTY) {
		err := pty.Stop()
		if err != nil {

		}
	}(pty)

	// 测试第一个命令
	var cmd1, cmd2 string
	var args1, args2 []string

	if runtime.GOOS == "windows" {
		cmd1 = "cmd"
		args1 = []string{"/c", "echo", "First Command"}
		cmd2 = "cmd"
		args2 = []string{"/c", "echo", "Second Command"}
	} else {
		cmd1 = "echo"
		args1 = []string{"First Command"}
		cmd2 = "echo"
		args2 = []string{"Second Command"}
	}

	// 启动第一个命令
	err := pty.Start(cmd1, args1, "")
	if err != nil {
		t.Fatalf("Failed to start first command: %v", err)
	}

	pty.ReadOutputToMock()
	time.Sleep(1 * time.Second)

	// 启动第二个命令
	err = pty.Start(cmd2, args2, "")
	if err != nil {
		t.Fatalf("Failed to start second command: %v", err)
	}

	pty.ReadOutputToMock()
	time.Sleep(1 * time.Second)

	// 检查消息
	messages := pty.GetMessages()
	if len(messages) == 0 {
		t.Fatal("No messages received")
	}

	// 应该能找到两个命令的输出
	foundFirst := false
	foundSecond := false
	for _, msg := range messages {
		if msg.Type == "stdout" {
			if strings.Contains(msg.Data, "First Command") {
				foundFirst = true
			}
			if strings.Contains(msg.Data, "Second Command") {
				foundSecond = true
			}
		}
	}

	if !foundFirst {
		t.Error("First command output not found")
	}
	if !foundSecond {
		t.Error("Second command output not found")
	}
}

func TestStandalonePTY_InputOutput(t *testing.T) {
	pty := NewStandalonePTY("test-session-3")
	defer func(pty *StandalonePTY) {
		err := pty.Stop()
		if err != nil {

		}
	}(pty)

	// 使用交互式命令进行测试
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{}
	} else {
		cmd = "sh"
		args = []string{}
	}

	// 启动交互式shell
	err := pty.Start(cmd, args, "")
	if err != nil {
		t.Fatalf("Failed to start interactive command: %v", err)
	}

	pty.ReadOutputToMock()
	time.Sleep(500 * time.Millisecond)

	// 发送输入
	testInput := "echo test input\n"
	if runtime.GOOS == "windows" {
		testInput = "echo test input\r\n"
	}

	err = pty.WriteInput(testInput)
	if err != nil {
		t.Fatalf("Failed to write input: %v", err)
	}

	// 等待输出
	time.Sleep(1 * time.Second)

	// 发送退出命令
	exitInput := "exit\n"
	if runtime.GOOS == "windows" {
		exitInput = "exit\r\n"
	}

	err = pty.WriteInput(exitInput)
	if err != nil {
		t.Logf("Failed to write exit input (this may be expected): %v", err)
	}

	time.Sleep(1 * time.Second)

	// 检查消息
	messages := pty.GetMessages()

	// 输出所有消息以便调试
	t.Logf("Received %d messages:", len(messages))
	for i, msg := range messages {
		t.Logf("Message %d: Type=%s, Data=%s, Status=%s, PID=%d", i, msg.Type, msg.Data, msg.Status, msg.PID)
	}

	// 至少应该有一些输出
	if len(messages) == 0 {
		t.Fatal("No messages received from interactive session")
	}
}

func TestStandalonePTY_ErrorCommand(t *testing.T) {
	pty := NewStandalonePTY("test-session-4")
	defer func(pty *StandalonePTY) {
		err := pty.Stop()
		if err != nil {

		}
	}(pty)

	// 尝试执行不存在的命令
	err := pty.Start("nonexistentcommand", []string{}, "")
	if err == nil {
		t.Fatal("Expected error when starting nonexistent command")
	}

	// 检查PTY状态
	if pty.Running {
		t.Error("PTY should not be running after failed start")
	}
}

func TestStandalonePTY_StopCommand(t *testing.T) {
	pty := NewStandalonePTY("test-session-5")

	// 启动长时间运行的命令
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "ping"
		args = []string{"127.0.0.1", "-n", "10"}
	} else {
		cmd = "sleep"
		args = []string{"5"}
	}

	err := pty.Start(cmd, args, "")
	if err != nil {
		t.Fatalf("Failed to start long-running command: %v", err)
	}

	pty.ReadOutputToMock()

	// 确认命令正在运行
	if !pty.Running {
		t.Error("PTY should be running")
	}

	// 等待一小段时间然后停止
	time.Sleep(1 * time.Second)

	err = pty.Stop()
	if err != nil {
		t.Errorf("Failed to stop PTY: %v", err)
	}

	// 确认命令已停止
	if pty.Running {
		t.Error("PTY should not be running after stop")
	}
}

func TestPTYManager(t *testing.T) {
	manager := GetPTYManager()

	// 创建测试PTY
	pty1 := NewStandalonePTY("test-manager-1")
	pty2 := NewStandalonePTY("test-manager-2")

	// 添加到管理器
	manager.AddPTY("test-1", pty1.PTY)
	manager.AddPTY("test-2", pty2.PTY)

	// 测试获取PTY
	retrievedPTY, exists := manager.GetPTY("test-1")
	if !exists {
		t.Error("PTY test-1 should exist in manager")
	}
	if retrievedPTY.ID != "test-manager-1" {
		t.Error("Retrieved PTY has wrong ID")
	}

	// 测试获取所有PTY
	allPTYs := manager.GetAllPTYs()
	if len(allPTYs) < 2 {
		t.Error("Should have at least 2 PTYs in manager")
	}

	// 测试移除PTY
	manager.RemovePTY("test-1")
	_, exists = manager.GetPTY("test-1")
	if exists {
		t.Error("PTY test-1 should not exist after removal")
	}

	// 清理
	manager.RemovePTY("test-2")
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"  ", nil},
		{"hello", []string{"hello"}},
		{"hello world", []string{"hello", "world"}},
		{`"hello world"`, []string{"hello world"}},
		{`hello "world test"`, []string{"hello", "world test"}},
		{`hello "world \"test\""`, []string{"hello", `world "test"`}},
		{`arg1 arg2 "arg with spaces"`, []string{"arg1", "arg2", "arg with spaces"}},
	}

	for _, test := range tests {
		result := parseArgs(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("For input %q, expected %v but got %v", test.input, test.expected, result)
			continue
		}

		for i, expected := range test.expected {
			if result[i] != expected {
				t.Errorf("For input %q, expected %v but got %v", test.input, test.expected, result)
				break
			}
		}
	}
}

// 基准测试
func BenchmarkPTYCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pty := NewStandalonePTY(fmt.Sprintf("bench-session-%d", i))
		_ = pty
	}
}

func BenchmarkPTYCommandExecution(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pty := NewStandalonePTY(fmt.Sprintf("bench-session-%d", i))

		var cmd string
		var args []string
		if runtime.GOOS == "windows" {
			cmd = "cmd"
			args = []string{"/c", "echo", "test"}
		} else {
			cmd = "echo"
			args = []string{"test"}
		}

		err := pty.Start(cmd, args, "")
		if err != nil {
			b.Fatalf("Failed to start command: %v", err)
		}

		pty.ReadOutputToMock()
		time.Sleep(100 * time.Millisecond)
		err = pty.Stop()
		if err != nil {
			return
		}
	}
}
