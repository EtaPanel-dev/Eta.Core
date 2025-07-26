package database

import (
	"golang.org/x/crypto/bcrypt"
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

	// 创建默认管理员用户
	d.createDefaultUser()

	return nil
}

func (d *Database) GetDb() *Database {
	if d.DbConn == nil {
		log.Fatal("数据库未初始化")
	}
	return d
}

// createDefaultUser 创建默认管理员用户
func (d *Database) createDefaultUser() {
	var count int64
	d.DbConn.Model(&models.User{}).Count(&count)

	// 如果没有用户，创建默认管理员
	if count == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Abc123456"), bcrypt.DefaultCost)
		if err != nil {
			log.Println("初始化用户失败，密码处理错误")
			return
		}
		defaultUser := models.User{
			Username: "demo",
			Password: string(hashedPassword),
		}

		if err := d.DbConn.Create(&defaultUser).Error; err != nil {
			log.Printf("创建默认用户失败: %v", err)
		} else {
			log.Println("已创建默认管理员用户 - 用户名: demo, 密码: Abc123456")
			log.Println("请尽快登录并修改默认密码！")
		}
	}
}

type Database struct {
	DbConn *gorm.DB
}
