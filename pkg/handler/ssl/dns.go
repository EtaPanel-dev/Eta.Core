package ssl

import (
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/database"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models/ssl"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CreateDnsAccount 创建 DNS 账号
func CreateDnsAccount(c *gin.Context) {
	var req ssl.CreateDnsAccountRequest
	var DbConn = database.DbConn

	if req.ProviderId < ssl.AliyunDnsProvider || req.ProviderId > ssl.BaiduCloudDnsProvider {
		handler.Respond(c, http.StatusBadRequest, "DNS 提供商不合法", nil)
		return
	}

	// 因为每个 dns 提供商所需要的参数可能不同，所以 key 传入的是 key1,key2,key3 value 传入的是 value1,value2,value3 一一对应
	dnsUser := &ssl.DnsUser{
		ProviderId: req.ProviderId,
		Key:        req.Key,
		Value:      req.Value,
	}

	if err := DbConn.Create(&dnsUser).Error; err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, nil, nil)
}

// GetDnsAccounts 获取所有 DNS 账号
func GetDnsAccounts(c *gin.Context) {
	var dnsUsers []ssl.DnsUser
	DbConn := database.DbConn

	if err := DbConn.Find(&dnsUsers).Error; err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var response []map[string]any
	if len(dnsUsers) < 1 {
		handler.Respond(c, http.StatusOK, nil, response)
		return
	}
	for _, user := range dnsUsers {
		var tmp map[string]any
		tmp["id"] = user.ProviderId
		tmp["key"] = user.Key
		tmp["value"] = user.Value
		response = append(response, tmp)
	}

	handler.Respond(c, http.StatusOK, nil, response)
}

// DeleteDnsAccount 删除 DNS 账号
func DeleteDnsAccount(c *gin.Context) {
	id := c.Param("id")
	var dnsUser ssl.DnsUser
	DbConn := database.DbConn

	if err := DbConn.First(&dnsUser, id).Error; err != nil {
		handler.Respond(c, http.StatusNotFound, "未找到 DNS 账号", nil)
		return
	}

	if err := DbConn.Delete(&dnsUser).Error; err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, nil, nil)
}

// UpdateDnsAccount 更新 DNS 账号
func UpdateDnsAccount(c *gin.Context) {
	id := c.Param("id")
	var req ssl.CreateDnsAccountRequest
	DbConn := database.DbConn

	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if req.ProviderId < ssl.AliyunDnsProvider || req.ProviderId > ssl.BaiduCloudDnsProvider {
		handler.Respond(c, http.StatusBadRequest, "DNS 提供商不合法", nil)
		return
	}

	var dnsUser ssl.DnsUser
	if err := DbConn.First(&dnsUser, id).Error; err != nil {
		handler.Respond(c, http.StatusNotFound, "未找到 DNS 账号", nil)
		return
	}

	dnsUser.ProviderId = req.ProviderId
	dnsUser.Key = req.Key
	dnsUser.Value = req.Value

	if err := DbConn.Save(&dnsUser).Error; err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, nil, nil)
}
