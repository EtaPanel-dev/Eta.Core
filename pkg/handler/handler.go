package handler

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Message string
	Code    int
}

func Respond(c *gin.Context, code int, message any, data interface{}) {
	res := gin.H{
		"status":  code,
		"message": message,
	}

	// 判断 data 是否为空，不为空则添加到响应中
	if data != nil {
		res["data"] = data
	}

	c.JSON(code, res)
}
