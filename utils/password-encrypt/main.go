package main

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
)

func main() {
	fmt.Print("请输入要加密的密码: ")
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("读取输入失败: %v\n", err)
		return
	}

	password = strings.TrimSpace(password)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("密码加密失败: %v\n", err)
		return
	}

	fmt.Printf("加密后的密码: %s\n", string(hashedPassword))
}
