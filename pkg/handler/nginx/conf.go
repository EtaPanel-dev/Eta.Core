package nginx

import (
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetNginxStatus 获取Nginx状态
func GetNginxStatus(c *gin.Context) {
	status := getNginxStatus()
	handler.Respond(c, http.StatusOK, "获取Nginx状态成功", status)
}

// GetNginxConfig 获取Nginx主配置
func GetNginxConfig(c *gin.Context) {
	config, err := getNginxMainConfig()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取Nginx配置失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "获取Nginx配置成功", config)
}

// UpdateNginxConfig 更新Nginx主配置
func UpdateNginxConfig(c *gin.Context) {
	var config models.NginxConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	if err := updateNginxMainConfig(config); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "更新Nginx配置失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "更新Nginx配置成功", nil)
}

// ResetNginxConfig 重置Nginx配置为默认
func ResetNginxConfig(c *gin.Context) {
	if err := resetNginxToDefault(); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "重置Nginx配置失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "重置Nginx配置成功", nil)
}
