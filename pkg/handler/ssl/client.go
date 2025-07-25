package ssl

import (
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/database"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models/ssl"
	ssl2 "github.com/EtaPanel-dev/Eta-Panel/core/pkg/ssl"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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
