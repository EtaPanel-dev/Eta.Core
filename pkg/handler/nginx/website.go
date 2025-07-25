package nginx

import (
	"net/http"
	"strconv"

	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models"
	"github.com/gin-gonic/gin"
)

// GetNginxSites 获取所有网站列表
// @Summary 获取网站列表
// @Description 获取所有Nginx网站配置列表
// @Tags Nginx管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=[]models.NginxSite} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/nginx/sites [get]
func GetNginxSites(c *gin.Context) {
	sites, err := getNginxSites()
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "获取网站列表失败: "+err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "获取网站列表成功", sites)
}

// CreateNginxSite 创建新网站
// @Summary 创建网站
// @Description 创建新的Nginx网站配置
// @Tags Nginx管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.NginxSite true "网站配置信息"
// @Success 200 {object} handler.Response "创建成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 409 {object} handler.Response "域名已存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/nginx/sites [post]
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
// @Summary 更新网站配置
// @Description 更新指定ID的网站配置
// @Tags Nginx管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "网站ID"
// @Param request body models.NginxSite true "网站配置信息"
// @Success 200 {object} handler.Response "更新成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 404 {object} handler.Response "网站不存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/nginx/sites/{id} [put]
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
// @Summary 删除网站
// @Description 删除指定ID的网站配置
// @Tags Nginx管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "网站ID"
// @Success 200 {object} handler.Response "删除成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 404 {object} handler.Response "网站不存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/nginx/sites/{id} [delete]
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
// @Summary 启用/禁用网站
// @Description 切换指定ID网站的启用状态
// @Tags Nginx管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "网站ID"
// @Success 200 {object} handler.Response "操作成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 404 {object} handler.Response "网站不存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/nginx/sites/{id}/toggle [post]
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
