package ssl

import (
	"github.com/LxHTT/Eta-Panel/core/pkg/models/ssl"
	"gorm.io/gorm"
)

func NewClient(db gorm.DB) *ssl.AcmeClient {
	acmeClient := &ssl.AcmeClient{
		Config: nil,
		Client: nil,
		User:   nil,
	}

	if err := db.First(&acmeClient).Error; err != nil {
		return nil
	}

	return acmeClient
}
