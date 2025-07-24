package middleware

import "github.com/gin-gonic/gin"

// Security 安全中间件
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 添加安全头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		c.Next()
	}
}
