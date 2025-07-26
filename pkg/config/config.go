package config

import (
	"log"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Server       ServerConfig    `toml:"server"`
	Database     DatabaseConfig  `toml:"database"`
	JWT          JWTConfig       `toml:"jwt"`
	IPFS         IPFSConfig      `toml:"ipfs"`
	Injective    InjectiveConfig `json:"injective" toml:"injective"` // Injective链配置
	DockerConfig DockerConfig    `json:"docker" toml:"docker"`       // Docker配置
}

type ServerConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

type DatabaseConfig struct {
	Path string `toml:"path"`
}

type JWTConfig struct {
	Secret string `toml:"secret"`
}

type IPFSConfig struct {
	URL     string `toml:"url"`
	Enabled bool   `toml:"enabled"`
}

type InjectiveConfig struct {
	Enabled     bool   `json:"enabled"`
	NetworkType string `json:"network_type"` // mainnet, testnet
	ChainID     string `json:"chain_id"`
	GRPCUrl     string `json:"grpc_url"`
	PrivateKey  string `json:"private_key"`
	GasPrice    string `json:"gas_price"`
}

// DockerConfig Docker配置
type DockerConfig struct {
	Enabled         bool             `json:"enabled"`
	SocketPath      string           `json:"socket_path"`
	APIVersion      string           `json:"api_version"`
	Registries      []DockerRegistry `json:"registries"`
	DefaultRegistry string           `json:"default_registry"`
}

// DockerRegistry Docker镜像仓库配置
type DockerRegistry struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Default  bool   `json:"default"`
}

// AppConfig 全局应用配置
var AppConfig *Config

func Init() error {
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
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
		JWT: JWTConfig{
			Secret: "your_jwt_secret",
		},
		IPFS: IPFSConfig{
			URL:     "http://localhost:5001",
			Enabled: true,
		},
		Injective: InjectiveConfig{
			Enabled:     false,
			NetworkType: "testnet",
			ChainID:     "injective-888",
			GRPCUrl:     "http://localhost:9090",
			PrivateKey:  "",
			GasPrice:    "0.025inj",
		},
		DockerConfig: DockerConfig{
			Enabled:    true,
			SocketPath: "/var/run/docker.sock",
			APIVersion: "1.41",
			Registries: []DockerRegistry{
				{
					Name:    "Docker Hub",
					URL:     "https://registry-1.docker.io",
					Default: true,
				},
				{
					Name: "阿里云镜像",
					URL:  "https://registry.cn-hangzhou.aliyuncs.com",
				},
				{
					Name: "网易云镜像",
					URL:  "https://hub-mirror.c.163.com",
				},
			},
			DefaultRegistry: "https://registry-1.docker.io",
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
