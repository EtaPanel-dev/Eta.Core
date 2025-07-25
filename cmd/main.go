package main

import (
	_ "github.com/EtaPanel-dev/Eta-Panel/core/cmd/api/docs"

	"fmt"
	"log"

	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/config"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/database"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/router"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title EtaPanel API
// @version 1.0
// @description EtaPanel 服务器管理面板后端API文档，提供系统监控、文件管理、Nginx配置、SSL证书管理等功能
// @termsOfService http://swagger.io/terms/

// @contact.name EtaPanel Support
// @contact.url https://github.com/EtaPanel-dev/Eta-Panel
// @contact.email support@etapanel.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入Bearer token，格式：Bearer {token}

// @tag.name 公共接口
// @tag.description 无需认证的公共接口

// @tag.name 认证
// @tag.description 用户认证相关接口

// @tag.name 系统监控
// @tag.description 系统资源监控和进程管理

// @tag.name 文件管理
// @tag.description 文件和目录操作管理

// @tag.name 定时任务
// @tag.description Crontab定时任务管理

// @tag.name Nginx管理
// @tag.description Nginx服务器配置和网站管理

// @tag.name SSL证书管理
// @tag.description SSL证书申请和管理

// @tag.name ACME客户端管理
// @tag.description ACME客户端配置管理

// @tag.name DNS账号管理
// @tag.description DNS服务商账号配置管理

func main() {

	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	db := database.InitDb()
	if err := db.Connect(); err != nil {
		log.Fatalf("database connection error: %s", err)
	}

	// 设置Gin模式
	gin.SetMode(gin.DebugMode)

	// 创建Gin引擎
	r := gin.Default()

	// 加载路由
	router.LoadRoutes(r)
	r.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	fmt.Printf("Server starting on %s\n", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
