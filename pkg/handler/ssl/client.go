package ssl

import (
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/database"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models/ssl"
	ssl2 "github.com/EtaPanel-dev/Eta-Panel/core/pkg/ssl"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// GetAcmeClients 获取所有 Acme 客户端
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
func DeleteAcmeClient(c *gin.Context) {
	id := c.Param("id")

	DbConn := database.DbConn
	if err := DbConn.Delete(&ssl.AcmeClient{}, id).Error; err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "Acme 客户端已删除", nil)
}
