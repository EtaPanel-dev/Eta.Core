package ssl

import "gorm.io/gorm"

type Ssl struct {
	gorm.Model
	Domain     string `gorm:"not null" json:"domain"`
	PublicKey  string `gorm:"unique" json:"certificate"`
	PrivateKey string `gorm:"not null" json:"key"`
	Enabled    bool   `gorm:"default:false" json:"enabled"`
	AutoRenew  bool   `gorm:"default:false" json:"auto_renew"`
	Email      string `gorm:"not null" json:"email"`
	Provider   string `gorm:"not null" json:"provider"`
}
