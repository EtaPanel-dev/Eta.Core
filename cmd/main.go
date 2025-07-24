package main

import (
	"fmt"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/config"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/database"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/router"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {

	db := database.InitDb()
	if err := db.Connect(); err != nil {
		log.Fatalf("database connection error: %s", err)
	}
	
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.Default()

	// 加载路由
	router.LoadRoutes(r)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	fmt.Printf("Server starting on %s\n", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
