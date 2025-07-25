package cmd

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type CommandHelper struct {
	workDir      string
	outputFile   string
	scriptPath   string
	timeout      time.Duration
	logger       *log.Logger
	IgnoreExist1 bool
}

type Option func(*CommandHelper)

func NewCommandMgr(opts ...Option) *CommandHelper {
	s := &CommandHelper{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func RunDefaultBashC(command string) error {
	mgr := NewCommandMgr()
	return mgr.RunBashC(command)
}

func RunDefaultBashCf(command string, arg ...interface{}) error {
	mgr := NewCommandMgr()
	return mgr.RunBashCf(command, arg...)
}

func RunDefaultWithStdoutBashC(command string) (string, error) {
	mgr := NewCommandMgr(WithTimeout(20 * time.Second))
	return mgr.RunWithStdoutBashC(command)
}

func (c *CommandHelper) Run(name string, arg ...string) error {
	_, err := c.run(name, arg...)
	return err
}

func (c *CommandHelper) RunBashCWithArgs(arg ...string) error {
	arg = append([]string{"-c"}, arg...)
	_, err := c.run("bash", arg...)
	return err
}

func (c *CommandHelper) RunBashC(command string) error {
	_, err := c.run("bash", "-c", command)
	return err
}

func (c *CommandHelper) RunBashCf(command string, arg ...interface{}) error {
	_, err := c.run("bash", "-c", fmt.Sprintf(command, arg...))
	return err
}

func (c *CommandHelper) RunWithStdout(name string, arg ...string) (string, error) {
	return c.run(name, arg...)
}

func (c *CommandHelper) RunWithStdoutBashC(command string) (string, error) {
	return c.run("bash", "-c", command)
}

func (c *CommandHelper) RunWithStdoutBashCf(command string, arg ...interface{}) (string, error) {
	return c.run("bash", "-c", fmt.Sprintf(command, arg...))
}

func (c *CommandHelper) run(name string, arg ...string) (string, error) {
	var cmd *exec.Cmd
	var ctx context.Context
	var cancel context.CancelFunc

	if c.timeout != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, name, arg...)
	} else {
		cmd = exec.Command(name, arg...)
	}

	var stdout, stderr bytes.Buffer
	if c.logger != nil {
		cmd.Stdout = c.logger.Writer()
		cmd.Stderr = c.logger.Writer()
	} else if len(c.outputFile) != 0 {
		file, err := os.OpenFile(c.outputFile, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return "", err
		}
		defer file.Close()
		cmd.Stdout = file
		cmd.Stderr = file
	} else if len(c.scriptPath) != 0 {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmd = exec.Command("bash", c.scriptPath)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	env := os.Environ()
	cmd.Env = env
	if len(c.workDir) != 0 {
		cmd.Dir = c.workDir
	}

	err := cmd.Run()

	if c.timeout != 0 && ctx != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("command timeout")
	}

	if err != nil {
		return handleErr(stdout, stderr, c.IgnoreExist1, err)
	}
	return stdout.String(), nil
}

func WithOutputFile(outputFile string) Option {
	return func(s *CommandHelper) {
		s.outputFile = outputFile
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(s *CommandHelper) {
		s.timeout = timeout
	}
}

func WithLogger(logger *log.Logger) Option {
	return func(s *CommandHelper) {
		s.logger = logger
	}
}

func WithWorkDir(workDir string) Option {
	return func(s *CommandHelper) {
		s.workDir = workDir
	}
}

func WithScriptPath(scriptPath string) Option {
	return func(s *CommandHelper) {
		s.scriptPath = scriptPath
	}
}

func WithIgnoreExist1() Option {
	return func(s *CommandHelper) {
		s.IgnoreExist1 = true
	}
}

func handleErr(stdout, stderr bytes.Buffer, ignoreExist1 bool, err error) (string, error) {
	var exitError *exec.ExitError
	if ignoreExist1 && errors.As(err, &exitError) {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			if status.ExitStatus() == 1 {
				return "", nil
			}
		}
	}
	errMsg := ""
	if len(stderr.String()) != 0 {
		errMsg = fmt.Sprintf("stderr: %s", stderr.String())
	}
	if len(stdout.String()) != 0 {
		if len(errMsg) != 0 {
			errMsg = fmt.Sprintf("%s; stdout: %s", errMsg, stdout.String())
		} else {
			errMsg = fmt.Sprintf("stdout: %s", stdout.String())
		}
	}
	return errMsg, err
}

// SudoHandleCmd 检测当前用户是否具有 sudo 权限
func SudoHandleCmd() string {
	cmd := exec.Command("sudo", "-n", "ls")
	if err := cmd.Run(); err == nil {
		return "sudo "
	}
	return ""
}

// Which 检查命令是否存在
func Which(name string) bool {
	stdout, err := RunDefaultWithStdoutBashCf("which %s", name)
	if err != nil || (len(strings.ReplaceAll(stdout, "\n", "")) == 0) {
		return false
	}
	return true
}

// ExecWithStreamOutput 实时执行命令并处理输出
func ExecWithStreamOutput(command string, outputCallback func(string)) error {
	cmd := exec.Command("bash", "-c", command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	go streamReader(stdout, outputCallback)
	go streamReader(stderr, outputCallback)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command finished with error: %w", err)
	}
	return nil
}

func streamReader(reader io.ReadCloser, callback func(string)) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		callback(scanner.Text())
	}
}

// RunDefaultWithStdoutBashCf executes a bash command with formatted arguments and returns stdout
func RunDefaultWithStdoutBashCf(format string, args ...interface{}) (string, error) {
	mgr := NewCommandMgr(WithTimeout(20 * time.Second))
	return mgr.RunWithStdoutBashCf(format, args...)
}
