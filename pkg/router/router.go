package router

import (
	"github.com/LxHTT/Eta-Panel/core/pkg/extend/pty"
	"github.com/LxHTT/Eta-Panel/core/pkg/handler/file"
	"github.com/LxHTT/Eta-Panel/core/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func LoadRoutes(r *gin.Engine) {
	// 初始化PTY服务
	pty.InitPTYService()

	// 添加中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Security())

	// 公共 API
	apiPublicRouter := r.Group("/api/public")
	{
		apiPublicRouter.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"code": 200, "message": "Eta Panel API Server Is OK!"})
		})
	}

	// 授权api
	apiAuthRouter := r.Group("/api/auth")
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
	}

	// 注册WebSocket路由 - 这里使用了PTY handlers中的函数
	r.GET("/ws/pty", gin.WrapH(pty.RegisterPTYHandler("/pty")))

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "message": "API route not found"})
	})
}
