package models

// Config 应用配置
type Config struct {
	IPFS      IPFSConfig      `json:"ipfs"`
	Injective InjectiveConfig `json:"injective"`
	Server    ServerConfig    `json:"server"`
}

// IPFSConfig IPFS配置
type IPFSConfig struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
	Timeout int    `json:"timeout"`
}

// InjectiveConfig Injective配置
type InjectiveConfig struct {
	Enabled     bool   `json:"enabled"`
	NetworkType string `json:"network_type"` // mainnet, testnet
	ChainID     string `json:"chain_id"`
	GRPCUrl     string `json:"grpc_url"`
	PrivateKey  string `json:"private_key"`
	GasPrice    string `json:"gas_price"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port        string `json:"port"`
	LogLevel    string `json:"log_level"`
	EnableHTTPS bool   `json:"enable_https"`
	CertFile    string `json:"cert_file"`
	KeyFile     string `json:"key_file"`
}
