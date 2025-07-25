package router

import (
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/extend/pty"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler/auth"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler/crontab"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler/file"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler/nginx"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler/ssl"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler/system"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func LoadRoutes(r *gin.Engine) {
	// 初始化PTY服务
	pty.InitPTYService()

	// 添加中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Security())

	// API版本控制
	v1 := r.Group("/api/v1")

	// 公共 API
	apiPublicRouter := r.Group("/api/public")
	{
		// @Summary 健康检查
		// @Description API服务器状态检查
		// @Tags 公共接口
		// @Accept json
		// @Produce json
		// @Success 200 {object} object{code=int,message=string} "服务正常"
		// @Router /api/public [get]
		apiPublicRouter.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"code": 200, "message": "Eta Panel API Server Is OK!"})
		})
		apiPublicRouter.POST("/login", auth.Login)
	}

	// 授权api
	apiAuthRouter := r.Group("/api/auth")
	apiAuthRouter.Use(middleware.JWTAuth()) // 添加JWT认证中间件
	{
		apiFileRouter := apiAuthRouter.Group("/files")
		{
			apiFileRouter.GET("/", file.ListFiles)
			apiFileRouter.GET("/download", file.DownloadFile)
			apiFileRouter.POST("/upload", file.UploadFile)
			apiFileRouter.POST("/move", file.MoveFile)
			apiFileRouter.POST("/copy", file.CopyFile)
			apiFileRouter.DELETE("/", file.DeleteFile)
			apiFileRouter.POST("/mkdir", file.CreateDirectory)
			apiFileRouter.POST("/compress", file.CompressFiles)
			apiFileRouter.POST("/extract", file.ExtractFiles)
			apiFileRouter.GET("/permissions", file.GetPermissions)
			apiFileRouter.POST("/permissions", file.SetPermissions)
			apiFileRouter.GET("/content", file.GetFileContent)
			apiFileRouter.POST("/content", file.SaveFileContent)
		}
		apiSysRouter := apiAuthRouter.Group("/system")
		{
			// 系统监控
			apiSysRouter.GET("/", system.GetSystemInfo)
			apiSysRouter.GET("/cpu", system.GetCPUInfo)
			apiSysRouter.GET("/memory", system.GetMemoryInfo)
			apiSysRouter.GET("/disk", system.GetDiskInfo)
			apiSysRouter.GET("/network", system.GetNetworkInfo)
			apiSysRouter.GET("/processes", system.GetProcessList)
			apiSysRouter.POST("/process/kill", system.KillProcess)
		}
		apiCronRouter := apiAuthRouter.Group("/crontab")
		{
			apiCronRouter.GET("/", crontab.GetCrontabList)
			apiCronRouter.POST("/", crontab.CreateCrontabEntry)
			apiCronRouter.PUT("/:id", crontab.UpdateCrontabEntry)
			apiCronRouter.DELETE("/:id", crontab.DeleteCrontabEntry)
			apiCronRouter.POST("/:id/toggle", crontab.ToggleCrontabEntry)
		}
		apiSslRouter := apiAuthRouter.Group("/acme/ssl")
		{
			apiSslRouter.GET("/", ssl.GetSSL)
			apiSslRouter.POST("/", ssl.IssueSSL)
			apiSslRouter.DELETE("/:id", ssl.DeleteSSL)
			apiSslRouter.PUT("/:id", ssl.UpdateSSL)
		}
		apiSslClientRouter := apiAuthRouter.Group("/acme/clients")
		{
			apiSslClientRouter.GET("/", ssl.GetAcmeClients)
			apiSslClientRouter.POST("/", ssl.CreateAcmeClient)
			apiSslClientRouter.PUT("/:id", ssl.UpdateAcmeClient)
			apiSslClientRouter.DELETE("/:id", ssl.DeleteAcmeClient)
		}
		apiSslDnsRouter := apiAuthRouter.Group("/acme/dns")
		{
			apiSslDnsRouter.GET("/", ssl.GetDnsAccounts)
			apiSslDnsRouter.POST("/", ssl.CreateDnsAccount)
			apiSslDnsRouter.PUT("/:id", ssl.UpdateDnsAccount)
			apiSslDnsRouter.DELETE("/:id", ssl.DeleteDnsAccount)
		}
		apiNginxRouter := apiAuthRouter.Group("/nginx")
		{
			// Nginx管理
			apiNginxRouter.GET("/status", nginx.GetNginxStatus)
			apiNginxRouter.GET("/config", nginx.GetNginxConfig)
			apiNginxRouter.PUT("/config", nginx.UpdateNginxConfig)
			apiNginxRouter.POST("/config/reset", nginx.ResetNginxConfig)
			apiNginxRouter.GET("/sites", nginx.GetNginxSites)
			apiNginxRouter.POST("/sites", nginx.CreateNginxSite)
			apiNginxRouter.PUT("/sites/:id", nginx.UpdateNginxSite)
			apiNginxRouter.DELETE("/sites/:id", nginx.DeleteNginxSite)
			apiNginxRouter.POST("/sites/:id/toggle", nginx.ToggleNginxSite)
			apiNginxRouter.POST("/restart", nginx.RestartNginx)
			apiNginxRouter.POST("/reload", nginx.ReloadNginx)
			apiNginxRouter.POST("/test", nginx.TestNginxConfig)
		}
	}

	// 注册WebSocket路由 - 这里使用了PTY handlers中的函数
	r.GET("/ws/pty", gin.WrapH(pty.RegisterPTYHandler("/pty")))

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "API route not found"})
	})
}
