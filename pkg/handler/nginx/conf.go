package nginx

import (
	"net/http"

	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models"
	"github.com/gin-gonic/gin"
)

// GetNginxStatus 获取Nginx状态
// @Summary 获取Nginx状态
// @Description 获取Nginx服务的运行状态信息
// @Tags Nginx管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=models.NginxStatus} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/nginx/status [get]
func GetNginxStatus(c *gin.Context) {
	status := getNginxStatus()
	handler.Respond(c, http.StatusOK, "获取Nginx状态成功", status)
}

// GetNginxConfig 获取Nginx主配置
// @Summary 获取Nginx主配置
// @Description 获取Nginx的主配置文件内容
// @Tags Nginx管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=models.NginxConfig} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/nginx/config [get]
func GetNginxConfig(c *gin.Context) {
	config, err := getNginxMainConfig()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取Nginx配置失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "获取Nginx配置成功", config)
}

// UpdateNginxConfig 更新Nginx主配置
// @Summary 更新Nginx主配置
// @Description 更新Nginx的主配置文件
// @Tags Nginx管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.NginxConfig true "Nginx配置信息"
// @Success 200 {object} handler.Response "更新成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/nginx/config [put]
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
// @Summary 重置Nginx配置
// @Description 将Nginx配置重置为默认设置
// @Tags Nginx管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response "重置成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/nginx/config/reset [post]
func ResetNginxConfig(c *gin.Context) {
	if err := resetNginxToDefault(); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "重置Nginx配置失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "重置Nginx配置成功", nil)
}
