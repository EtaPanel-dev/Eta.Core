package database

import (
	"fmt"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/config"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models/ssl"
	"gorm.io/driver/mysql"
	"log"

	"gorm.io/gorm"
)

var DbConn *gorm.DB

func InitDb() *Database {
	Db := new(Database)
	return Db
}

func (d *Database) Connect() error {
	dbConfig := config.AppConfig.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Pwd,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Database)

	DbC, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

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
