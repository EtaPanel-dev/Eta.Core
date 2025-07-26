package auth

import (
	"net/http"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/database"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models"
	"github.com/gin-gonic/gin"
)

// ChangePasswordRequest 修改密码请求参数
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param passwordData body ChangePasswordRequest true "密码修改信息"
// @Success 200 {object} handler.Response "密码修改成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "旧密码错误"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/change-password [post]
// @Security BearerAuth
func ChangePassword(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		handler.Respond(c, http.StatusUnauthorized, "未找到用户信息", nil)
		return
	}

	var passwordData ChangePasswordRequest
	if err := c.ShouldBindJSON(&passwordData); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误: "+err.Error(), nil)
		return
	}

	// 获取当前用户
	var user models.User
	if err := database.DbConn.Where("username = ?", username).First(&user).Error; err != nil {
		handler.Respond(c, http.StatusUnauthorized, "用户不存在", nil)
		return
	}

	// 验证旧密码
	if err := user.CheckPassword(passwordData.OldPassword); err != nil {
		handler.Respond(c, http.StatusUnauthorized, "旧密码错误", nil)
		return
	}

	// 更新密码
	user.Password = passwordData.NewPassword
	if err := user.HashPassword(); err != nil {
		handler.Respond(c, http.StatusInternalServerError, "密码加密失败", nil)
		return
	}

	if err := database.DbConn.Save(&user).Error; err != nil {
		handler.Respond(c, http.StatusInternalServerError, "密码更新失败", nil)
		return
	}

	handler.Respond(c, http.StatusOK, "密码修改成功", nil)
}
