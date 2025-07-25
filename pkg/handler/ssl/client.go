package ssl

import (
	"net/http"
	"time"

	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/database"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models/ssl"
	ssl2 "github.com/EtaPanel-dev/Eta-Panel/core/pkg/ssl"
	"github.com/gin-gonic/gin"
)

// GetAcmeClients 获取所有 Acme 客户端
// @Summary 获取ACME客户端列表
// @Description 获取所有ACME客户端配置
// @Tags ACME客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=[]ssl.AcmeClientResponse} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/acme/clients [get]
func GetAcmeClients(c *gin.Context) {
	var clients []ssl.AcmeClient
	DbConn := database.DbConn
	if err := DbConn.Find(&clients).Error; err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response := make([]ssl.AcmeClientResponse, len(clients))
	for i, client := range clients {
		response[i] = ssl.AcmeClientResponse{
			Id:        client.ID,
			CreatedAt: client.CreatedAt,
			Email:     client.User.Email,
			ServerURL: client.ServerURL,
		}
	}

	handler.Respond(c, http.StatusOK, nil, response)
}

// GetAcmeClient 获取单个 ACME 客户端
func GetAcmeClient(id int) (ssl.AcmeClient, error) {

	var client ssl.AcmeClient
	DbConn := database.DbConn
	if err := DbConn.First(&client, id).Error; err != nil {
		return client, err
	}

	return client, nil
}

// CreateAcmeClient 创建 ACME 客户端
// @Summary 创建ACME客户端
// @Description 创建新的ACME客户端配置
// @Tags ACME客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ssl.CreateAcmeClientRequest true "ACME客户端信息"
// @Success 200 {object} handler.Response "创建成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/acme/clients [post]
func CreateAcmeClient(c *gin.Context) {
	var req ssl.CreateAcmeClientRequest
	DbConn := database.DbConn
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	baseModel := ssl.BaseModel{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	webSiteAcmeAccount := &ssl.WebsiteAcmeAccount{
		BaseModel: baseModel,
		Email:     req.Email,
		CaDirURL:  ssl2.GetCaDirURL(req.KeyType, ""),
		KeyType:   req.KeyType,
	}

	client, err := ssl2.NewAcmeClient(webSiteAcmeAccount)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
	}

	if err := DbConn.Create(&client).Error; err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusCreated, nil, ssl.AcmeClientResponse{
		Id:        client.ID,
		CreatedAt: client.CreatedAt,
		Email:     client.User.Email,
		ServerURL: client.ServerURL,
	})
}

// UpdateAcmeClient 更新ACME客户端
// @Summary 更新ACME客户端
// @Description 更新指定ID的ACME客户端配置
// @Tags ACME客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Param request body ssl.CreateAcmeClientRequest true "ACME客户端信息"
// @Success 200 {object} handler.Response "更新成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 404 {object} handler.Response "客户端不存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/acme/clients/{id} [put]
func UpdateAcmeClient(c *gin.Context) {
	id := c.Param("id")

	var req ssl.UpdateAcmeClientRequest
	DbConn := database.DbConn
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var client ssl.AcmeClient
	if err := DbConn.First(&client, id).Error; err != nil {
		handler.Respond(c, http.StatusNotFound, "未找到 Acme 客户端", nil)
		return
	}

	client.ServerURL = req.ServerURL

	if err := DbConn.Save(&client).Error; err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, nil, ssl.AcmeClientResponse{
		Id:        client.ID,
		CreatedAt: client.CreatedAt,
	})
}

// DeleteAcmeClient 删除 ACME 客户端
// @Summary 删除ACME客户端
// @Description 删除指定ID的ACME客户端配置
// @Tags ACME客户端管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户端ID"
// @Success 200 {object} handler.Response "删除成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 404 {object} handler.Response "客户端不存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /api/auth/acme/clients/{id} [delete]
func DeleteAcmeClient(c *gin.Context) {
	id := c.Param("id")

	DbConn := database.DbConn
	if err := DbConn.Delete(&ssl.AcmeClient{}, id).Error; err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "Acme 客户端已删除", nil)
}
