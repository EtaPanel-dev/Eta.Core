package middleware

import (
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请求头中缺少Authorization字段",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		//parts := strings.SplitN(authHeader, " ", 2)
		//if !(len(parts) == 2 && parts[0] == "Bearer") {
		//	c.JSON(http.StatusUnauthorized, gin.H{
		//		"code":    401,
		//		"message": "Authorization格式错误",
		//	})
		//	c.Abort()
		//	return
		//}

		// 解析token，使用配置文件中的密钥
		jwtSecret := []byte(config.AppConfig.JWT.Secret)
		token, err := jwt.ParseWithClaims(authHeader, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token解析失败: " + err.Error(),
			})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			// 将用户信息存储到上下文中
			c.Set("username", claims.Username)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token无效",
			})
			c.Abort()
			return
		}
	}
}
