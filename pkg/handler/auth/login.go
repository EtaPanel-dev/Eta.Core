package auth

import (
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/config"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"demo" `
	Password string `json:"password" binding:"required" example:"Abc123456" `
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNjQwOTk1MjAwfQ.signature"`
	ExpiresAt int64  `json:"expires_at" example:"1640995200"`
}

// Login 登录
// @Summary 用户登录
// @Description 通过用户名和密码进行登录，返回JWT token
// @Tags 认证
// @Accept json
// @Produce json
// @Param loginData body LoginRequest true "登录信息"
// @Success 200 {object} handler.Response{data=LoginResponse} "登录成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "用户名或密码错误"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/public/login [post]
func Login(c *gin.Context) {

	var loginData LoginRequest
	if err := c.ShouldBindJSON(&loginData); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
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
	jwtSecret := []byte(config.AppConfig.JWT.Secret)
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
