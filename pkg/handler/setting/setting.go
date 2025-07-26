package setting

import (
	"net/http"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/setting"
	"github.com/gin-gonic/gin"
)

// GetSettings 做一个 gin 的路由函数执行获取设置项的操作以 json 输出
func GetSettings(c *gin.Context) {
	// 获取设置项
	settings, err := setting.GetPanelSettings()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取设置失败: "+err.Error(), nil)
		return
	}

	// 返回设置项
	handler.Respond(c, http.StatusOK, "获取设置成功", settings)
}

// SaveSettings 做一个 gin 的路由函数执行保存设置项的操作
func SaveSettings(c *gin.Context) {
	var settings models.PanelSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	// 保存设置项
	if err := setting.SavePanelSettings(settings); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "保存设置失败: "+err.Error(), nil)
		return
	}

	// 返回成功响应
	handler.Respond(c, http.StatusOK, "保存设置成功", nil)
}
