package ssl

import (
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/database"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models/ssl"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetSSL handles the request to retrieve SSL data from the database.
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
