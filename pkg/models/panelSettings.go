package models

type PanelSettings struct {
	Name       string `json:"name" binding:"required"`        // 面板名称
	Logo       string `json:"logo" binding:"required"`        // Logo URL
	ThemeColor string `json:"theme_color" binding:"required"` // 主题色
}
