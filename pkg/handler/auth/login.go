package auth

import (
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-secret-key") // 在生产环境中应该从配置文件读取

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func Login(c *gin.Context) {

	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		handler.Respond(c, http.StatusBadRequest, nil, nil)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte("$2a$10$..."), []byte(loginData.Password)) != nil {
		handler.Respond(c, http.StatusUnauthorized, "密码错误", 401)
		return
	}

	// 生成JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: loginData.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "生成密钥失败", nil)
		return
	}

	handler.Respond(c, http.StatusOK, gin.H{
		"token":      tokenString,
		"expires_at": expirationTime.Unix(),
	}, "登录成功")
}
