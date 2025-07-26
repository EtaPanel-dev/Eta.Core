package router

import (
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/extend/pty"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/ai"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/auth"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/crontab"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/docker"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/file"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/firewall"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/log"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/nginx"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/setting"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/ssl"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler/system"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func LoadRoutes(r *gin.Engine) {
	// 初始化PTY服务
	pty.InitPTYService()

	// 添加中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Security())
	r.Use(middleware.LogVerification())

	// 公共 API
	apiPublicRouter := r.Group("/api/public")
	{
		// @Summary 健康检查
		// @Description API服务器状态检查
		// @Tags 公共接口
		// @Accept json
		// @Produce json
		// @Success 200 {object} object{code=int,message=string} "服务正常"
		// @Router /public [get]
		apiPublicRouter.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{"code": 200, "message": "Eta Panel API Server Is OK!"})
		})
		apiPublicRouter.POST("/login", auth.Login)
	}

	// 授权API
	apiAuthRouter := r.Group("/api/auth")
	//apiAuthRouter.Use(middleware.JWTAuth()) // 添加JWT认证中间件
	{
		apiAuthRouter.POST("/change-password", auth.ChangePassword)

		// 文件管理API
		apiFileRouter := apiAuthRouter.Group("/files")
		{
			apiFileRouter.GET("", file.ListFiles)
			apiFileRouter.GET("/download", file.DownloadFile)
			apiFileRouter.POST("/upload", file.UploadFile)
			apiFileRouter.POST("/move", file.MoveFile)
			apiFileRouter.POST("/copy", file.CopyFile)
			apiFileRouter.DELETE("", file.DeleteFile)
			apiFileRouter.POST("/mkdir", file.CreateDirectory)
			apiFileRouter.POST("/compress", file.CompressFiles)
			apiFileRouter.POST("/extract", file.ExtractFiles)
			apiFileRouter.GET("/permissions", file.GetPermissions)
			apiFileRouter.POST("/permissions", file.SetPermissions)
			apiFileRouter.GET("/content", file.GetFileContent)
			apiFileRouter.POST("/content", file.SaveFileContent)
		}

		// 系统监控API
		apiSysRouter := apiAuthRouter.Group("/system")
		{
			apiSysRouter.GET("", system.GetSystemInfo)
			apiSysRouter.GET("/cpu", system.GetCPUInfo)
			apiSysRouter.GET("/memory", system.GetMemoryInfo)
			apiSysRouter.GET("/disk", system.GetDiskInfo)
			apiSysRouter.GET("/network", system.GetNetworkInfo)
			apiSysRouter.GET("/processes", system.GetProcessList)
			apiSysRouter.POST("/process/kill", system.KillProcess)
		}

		// 定时任务API
		apiCronRouter := apiAuthRouter.Group("/crontab")
		{
			apiCronRouter.GET("", crontab.GetCrontabList)
			apiCronRouter.POST("", crontab.CreateCrontabEntry)
			apiCronRouter.PUT("/:id", crontab.UpdateCrontabEntry)
			apiCronRouter.DELETE("/:id", crontab.DeleteCrontabEntry)
			apiCronRouter.POST("/:id/toggle", crontab.ToggleCrontabEntry)
		}

		// SSL证书管理API
		apiSslRouter := apiAuthRouter.Group("/acme/ssl")
		{
			apiSslRouter.GET("", ssl.GetSSL)
			apiSslRouter.POST("", ssl.IssueSSL)
			apiSslRouter.DELETE("/:id", ssl.DeleteSSL)
			apiSslRouter.PUT("/:id", ssl.UpdateSSL)
		}

		// ACME客户端管理API
		apiSslClientRouter := apiAuthRouter.Group("/acme/clients")
		{
			apiSslClientRouter.GET("", ssl.GetAcmeClients)
			apiSslClientRouter.POST("", ssl.CreateAcmeClient)
			apiSslClientRouter.PUT("/:id", ssl.UpdateAcmeClient)
			apiSslClientRouter.DELETE("/:id", ssl.DeleteAcmeClient)
		}

		// DNS账户管理API
		apiSslDnsRouter := apiAuthRouter.Group("/acme/dns")
		{
			apiSslDnsRouter.GET("", ssl.GetDnsAccounts)
			apiSslDnsRouter.POST("", ssl.CreateDnsAccount)
			apiSslDnsRouter.PUT("/:id", ssl.UpdateDnsAccount)
			apiSslDnsRouter.DELETE("/:id", ssl.DeleteDnsAccount)
		}

		// Nginx管理API
		apiNginxRouter := apiAuthRouter.Group("/nginx")
		{
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

		// AI工具链API
		apiAiRouter := apiAuthRouter.Group("/ai")
		{
			apiAiRouter.POST("/log", ai.AnalyzeLog)
			apiAiRouter.POST("/files", ai.AnalyzeFiles)
		}

		// 系统设置API
		apiSettingRouter := apiAuthRouter.Group("/setting")
		{
			apiSettingRouter.PUT("", setting.SaveSettings)
			apiSettingRouter.GET("", setting.GetSettings)
		}

		// 防火墙管理API
		apiFirewallRouter := apiAuthRouter.Group("/firewall")
		{
			apiFirewallRouter.GET("/status", firewall.GetFirewallStatus)
			apiFirewallRouter.POST("/enable", firewall.EnableFirewall)
			apiFirewallRouter.POST("/disable", firewall.DisableFirewall)
			apiFirewallRouter.POST("/rules", firewall.AddFirewallRule)
			apiFirewallRouter.DELETE("/rules/:id", firewall.DeleteFirewallRule)
			apiFirewallRouter.POST("/reset", firewall.ResetFirewall)
			apiFirewallRouter.POST("/allow-ssh", firewall.AllowSSH)
			apiFirewallRouter.POST("/deny-all", firewall.DenyAll)
		}

		// 日志管理API
		apiLogRouter := r.Group("/log")
		{
			apiLogRouter.POST("/query", log.GetLogByRequestID)
			apiLogRouter.POST("/verify", log.VerifyLogIntegrity)
			apiLogRouter.GET("/list", log.ListLogHashes)
			apiLogRouter.GET("/stats", log.GetLogStats)
		}
		// Docker管理API
		apiDockerRouter := r.Group("/docker")
		{
			// 镜像管理
			apiDockerRouter.GET("/images", docker.GetDockerImages)
			apiDockerRouter.POST("/images/pull", docker.PullDockerImage)
			apiDockerRouter.DELETE("/images/:id", docker.DeleteDockerImage)

			// 容器管理
			apiDockerRouter.GET("/containers", docker.GetDockerContainers)
			apiDockerRouter.POST("/containers/run", docker.RunDockerContainer)
			apiDockerRouter.POST("/containers/:id/start", docker.StartDockerContainer)
			apiDockerRouter.POST("/containers/:id/stop", docker.StopDockerContainer)
			apiDockerRouter.DELETE("/containers/:id", docker.RemoveDockerContainer)

			// 终端连接
			apiDockerRouter.GET("/containers/:id/terminal", docker.DockerTerminal)

			// 镜像源管理
			apiDockerRouter.POST("/registry", docker.SetDockerRegistry)
		}

		// WebSocket连接（PTY终端）
		r.GET("/ws/pty", gin.WrapH(pty.RegisterPTYHandler("/pty")))
	}
	// 404错误处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"code":    404,
			"message": "API route not found",
			"error":   "The requested endpoint does not exist",
		})
	})
}
