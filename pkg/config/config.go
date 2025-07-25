package config

import (
	"log"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Server   ServerConfig   `toml:"server"`
	Database DatabaseConfig `toml:"database"`
}

type ServerConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

type DatabaseConfig struct {
	Path string `toml:"path"` // SQLite 数据库文件路径
}

var AppConfig *Config

func Init() error {
	// 检查配置文件是否存在
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置
		createDefaultConfig()
		log.Println("Created default config.toml file. Please modify it and restart the application.")
		os.Exit(0)
	}

	data, err := os.ReadFile("config.toml")
	if err != nil {
		return err
	}

	var cfg Config
	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}

	AppConfig = &cfg
	return nil
}

func createDefaultConfig() {
	defaultConfig := Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Path: "data.db",
		},
	}

	data, err := toml.Marshal(defaultConfig)
	if err != nil {
		log.Fatalf("Failed to marshal default config: %v", err)
	}

	err = os.WriteFile("config.toml", data, 0644)
	if err != nil {
		log.Fatalf("Failed to write default config file: %v", err)
	}

	log.Println("Default config.toml file created successfully, please modify it and restart the application.")
}
