package ssl

import (
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/database"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models/ssl"
)

func GetUserById(id int) (ssl.AcmeUser, error) {
	var user ssl.AcmeUser

	DbConn := database.DbConn
	if err := DbConn.First(&user, id).Error; err != nil {
		return user, err
	}
	return user, nil
}
