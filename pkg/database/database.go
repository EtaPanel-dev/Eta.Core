package database

import (
	"log"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/config"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models/ssl"
	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

var DbConn *gorm.DB

func InitDb() *Database {
	Db := new(Database)
	return Db
}

func (d *Database) Connect() error {
	dbConfig := config.AppConfig.Database

	DbC, err := gorm.Open(sqlite.Open(dbConfig.Path), &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动迁移数据库表
	err = DbC.AutoMigrate(
		&models.User{},
		&models.Server{},
		&models.AuthToken{},
		&ssl.AcmeClient{},
		&ssl.WebsiteAcmeAccount{},
		&ssl.AcmeUser{},
		&ssl.DnsUser{},
	)
	if err != nil {
		return err
	}

	d.DbConn = DbC
	DbConn = DbC

	return nil
}

func (d *Database) GetDb() *Database {
	if d.DbConn == nil {
		log.Fatal("数据库未初始化")
	}
	return d
}

type Database struct {
	DbConn *gorm.DB
}
