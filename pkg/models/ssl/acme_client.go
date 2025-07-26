package ssl

import (
	"github.com/go-acme/lego/v4/lego"
	"gorm.io/gorm"
	"time"
)

type AcmeClient struct {
	gorm.Model
	Id        uint
	Config    *lego.Config `gorm:"-"`
	Client    *lego.Client `gorm:"-"`
	ServerURL string
	User      *AcmeUser `gorm:"-"`
	KeyType   string
}

type CreateAcmeClientRequest struct {
	Email   string `json:"email" binding:"required,email"`
	KeyType string `json:"key_type" binding:"required"`
	Server  string `json:"server" binding:"required,server"`
}

type AcmeClientResponse struct {
	Id        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Email     string    `json:"email"`
	ServerURL string    `json:"server_url"`
}

type UpdateAcmeClientRequest struct {
	Id        uint   `json:"id"`
	ServerURL string `json:"server_url"`
}
