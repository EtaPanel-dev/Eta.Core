package ssl

import (
	"github.com/go-acme/lego/v4/lego"
	"gorm.io/gorm"
)

type AcmeClient struct {
	gorm.Model
	Config *lego.Config
	Client *lego.Client
	User   *AcmeUser
}
