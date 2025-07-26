package setting

import (
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/database"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models"
)

func GetPanelSettings() (models.PanelSettings, error) {
	DbConn := database.DbConn
	var settings models.PanelSettings
	err := DbConn.First(&settings).Error
	if err != nil {
		return settings, err
	}
	return settings, nil
}

func SavePanelSettings(settings models.PanelSettings) error {
	DbConn := database.DbConn
	// 检查是否已存在设置
	var existingSettings models.PanelSettings
	err := DbConn.First(&existingSettings).Error
	if err != nil {
		// 如果不存在，则创建新设置
		return DbConn.Create(&settings).Error
	}
	// 如果已存在，则更新设置
	return DbConn.Model(&existingSettings).Updates(settings).Error
}
