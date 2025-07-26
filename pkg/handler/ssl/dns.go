package ssl

import (
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/database"
	"net/http"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models/ssl"
	"github.com/gin-gonic/gin"
)

// CreateDnsAccount 创建 DNS 账号
// @Summary 创建DNS账号
// @Description 创建新的DNS服务商账号配置
// @Tags DNS账号管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ssl.CreateDnsAccountRequest true "DNS账号信息"
// @Success 200 {object} handler.Response "创建成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/acme/dns [post]
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
// @Summary 获取DNS账号列表
// @Description 获取所有DNS服务商账号配置
// @Tags DNS账号管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=[]object} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/acme/dns [get]
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
// @Summary 删除DNS账号
// @Description 删除指定ID的DNS服务商账号配置
// @Tags DNS账号管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "账号ID"
// @Success 200 {object} handler.Response "删除成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 404 {object} handler.Response "账号不存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/acme/dns/{id} [delete]
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
// @Summary 更新DNS账号
// @Description 更新指定ID的DNS服务商账号配置
// @Tags DNS账号管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "账号ID"
// @Param request body ssl.CreateDnsAccountRequest true "DNS账号信息"
// @Success 200 {object} handler.Response "更新成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 404 {object} handler.Response "账号不存在"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/acme/dns/{id} [put]
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
