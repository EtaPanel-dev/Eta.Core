package models

import (
	"gorm.io/gorm"
	"time"
)

type AuthToken struct {
	gorm.Model
	Id        uint      `json:"id" gorm:"primaryKey"`
	Token     string    `json:"token" gorm:"unique;not null"`
	Ua        string    `json:"ua" gorm:"unique;not null"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired 检查令牌是否过期
func (t *AuthToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
