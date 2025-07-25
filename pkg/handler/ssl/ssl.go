package ssl

import (
	"net/http"

	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/database"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models/ssl"
	"github.com/gin-gonic/gin"
)

// GetSSL handles the request to retrieve SSL data from the database.
// @Summary 获取SSL证书列表
// @Description 获取所有SSL证书信息
// @Tags SSL证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=[]object} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/acme/ssl [get]
func GetSSL(c *gin.Context) {
	// 从数据库获取ssl列表
	var sslList []ssl.Ssl
	db := database.DbConn
	if err := db.Find(&sslList).Error; err != nil {
		handler.Respond(c, http.StatusBadRequest, "Failed to retrieve SSL data", err.Error())
	}
	handler.Respond(c, http.StatusOK, nil, sslList)
}

// DeleteSSL handles the request to delete an SSL entry by ID.
// @Summary 删除SSL证书
// @Description 删除指定ID的SSL证书
// @Tags SSL证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "证书ID"
// @Success 200 {object} handler.Response "删除成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 404 {object} handler.Response "证书不存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/acme/ssl/{id} [delete]
func DeleteSSL(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		handler.Respond(c, http.StatusBadRequest, "SSL ID is required", nil)
		return
	}

	var sslEntry ssl.Ssl
	db := database.DbConn
	if err := db.First(&sslEntry, id).Error; err != nil {
		handler.Respond(c, http.StatusNotFound, "SSL entry not found", err.Error())
		return
	}

	if err := db.Delete(&sslEntry).Error; err != nil {
		handler.Respond(c, http.StatusInternalServerError, "Failed to delete SSL entry", err.Error())
		return
	}

	handler.Respond(c, http.StatusOK, "SSL entry deleted successfully", nil)
}

// UpdateSSL handles the request to update an existing SSL entry.
// @Summary 更新SSL证书
// @Description 更新指定ID的SSL证书信息
// @Tags SSL证书管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "证书ID"
// @Param request body object{domain=string,certificate=string,key=string,enabled=bool,auto_renew=bool,email=string,provider=string} true "SSL证书信息"
// @Success 200 {object} handler.Response "更新成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 404 {object} handler.Response "证书不存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/acme/ssl/{id} [put]
func UpdateSSL(c *gin.Context) {
	var sslEntry ssl.Ssl
	db := database.DbConn
	if err := c.ShouldBindJSON(&sslEntry); err != nil {
		handler.Respond(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	if err := db.Save(&sslEntry).Error; err != nil {
		handler.Respond(c, http.StatusInternalServerError, "Failed to update SSL entry", err.Error())
		return
	}

	handler.Respond(c, http.StatusOK, nil, sslEntry)
}
