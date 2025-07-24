package ssl

import (
	"crypto"
	"github.com/go-acme/lego/v4/registration"
	"gorm.io/gorm"
	"time"
)

type AcmeUser struct {
	gorm.Model
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}

func (u *AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}

type BaseModel struct {
	ID        uint      `gorm:"primarykey;AUTO_INCREMENT" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type WebsiteAcmeAccount struct {
	gorm.Model
	BaseModel
	Email      string `gorm:"not null" json:"email"`
	URL        string `gorm:"not null" json:"url"`
	PrivateKey string `gorm:"not null" json:"-"`
	Type       string `gorm:"not null;default:letsencrypt" json:"type"`
	EabKid     string `json:"eabKid"`
	EabHmacKey string `json:"eabHmacKey"`
	KeyType    string `gorm:"not null;default:2048" json:"keyType"`
	UseProxy   bool   `gorm:"default:false" json:"useProxy"`
	CaDirURL   string `json:"caDirURL"`
}
