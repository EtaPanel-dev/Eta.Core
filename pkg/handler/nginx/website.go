package nginx

import (
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetNginxSites 获取所有网站列表
func GetNginxSites(c *gin.Context) {
	sites, err := getNginxSites()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取网站列表失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "获取网站列表成功", sites)
}

// CreateNginxSite 创建新网站
func CreateNginxSite(c *gin.Context) {
	var site models.NginxSite
	if err := c.ShouldBindJSON(&site); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	// 验证必填字段
	if site.Name == "" || site.Domain == "" {
		handler.Respond(c, http.StatusBadRequest, "网站名称和域名不能为空", nil)
		return
	}

	// 检查域名是否已存在
	if exists, err := domainExists(site.Domain); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "检查域名失败: "+err.Error(), nil)
		return
	} else if exists {
		handler.Respond(c, http.StatusBadRequest, "域名已存在", nil)
		return
	}

	// 创建网站配置
	if err := createNginxSite(site); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "创建网站失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "创建网站成功", nil)
}

// UpdateNginxSite 更新网站配置
func UpdateNginxSite(c *gin.Context) {
	siteID := c.Param("id")
	id, err := strconv.Atoi(siteID)
	if err != nil {
		handler.Respond(c, http.StatusBadRequest, "无效的网站ID", nil)
		return
	}

	var site models.NginxSite
	if err := c.ShouldBindJSON(&site); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	site.ID = id

	if err := updateNginxSite(site); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "更新网站失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "更新网站成功", nil)
}

// DeleteNginxSite 删除网站
func DeleteNginxSite(c *gin.Context) {
	siteID := c.Param("id")
	id, err := strconv.Atoi(siteID)
	if err != nil {
		handler.Respond(c, http.StatusBadRequest, "无效的网站ID", nil)
		return
	}

	if err := deleteNginxSite(id); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "删除网站失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "删除网站成功", nil)
}

// ToggleNginxSite 启用/禁用网站
func ToggleNginxSite(c *gin.Context) {
	siteID := c.Param("id")
	id, err := strconv.Atoi(siteID)
	if err != nil {
		handler.Respond(c, http.StatusBadRequest, "无效的网站ID", nil)
		return
	}

	if err := toggleNginxSite(id); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "切换网站状态失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "切换网站状态成功", nil)
}
