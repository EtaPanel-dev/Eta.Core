package models

// NginxSite 网站配置结构
type NginxSite struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Domain     string   `json:"domain"`
	Aliases    []string `json:"aliases"`
	Root       string   `json:"root"`
	Index      string   `json:"index"`
	SSL        bool     `json:"ssl"`
	SSLCert    string   `json:"sslCert"`
	SSLKey     string   `json:"sslKey"`
	ForceHTTPS bool     `json:"forceHttps"`
	Proxy      bool     `json:"proxy"`
	ProxyPass  string   `json:"proxyPass"`
	Rewrite    string   `json:"rewrite"`
	AccessLog  string   `json:"accessLog"`
	ErrorLog   string   `json:"errorLog"`
	Enabled    bool     `json:"enabled"`
	ConfigPath string   `json:"configPath"`
	CreatedAt  string   `json:"createdAt"`
	UpdatedAt  string   `json:"updatedAt"`
}

// NginxConfig Nginx主配置
type NginxConfig struct {
	ConfigPath    string `json:"configPath"`
	User          string `json:"user"`
	WorkerProcess string `json:"workerProcess"`
	ErrorLog      string `json:"errorLog"`
	AccessLog     string `json:"accessLog"`
	PidFile       string `json:"pidFile"`
	WorkerConn    string `json:"workerConn"`
	Gzip          bool   `json:"gzip"`
	ServerTokens  bool   `json:"serverTokens"`
}

// NginxStatus Nginx状态
type NginxStatus struct {
	Running     bool   `json:"running"`
	PID         int    `json:"pid"`
	Version     string `json:"version"`
	ConfigTest  bool   `json:"configTest"`
	Uptime      string `json:"uptime"`
	Connections int    `json:"connections"`
}

const (
	NginxConfigPath     = "/etc/nginx/nginx.conf"
	NginxSitesAvailable = "/etc/nginx/sites-available"
	NginxSitesEnabled   = "/etc/nginx/sites-enabled"
	DefaultNginxConfig  = `user www-data;
worker_processes auto;
pid /run/nginx.pid;
include /etc/nginx/modules-enabled/*.conf;

events {
    worker_connections 768;
    # multi_accept on;
}

http {
    ##
    # Basic Settings
    ##

    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    server_tokens off;

    # server_names_hash_bucket_size 64;
    # server_name_in_redirect off;

    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    ##
    # SSL Settings
    ##

    ssl_protocols TLSv1 TLSv1.1 TLSv1.2 TLSv1.3; # Dropping SSLv3, ref: POODLE
    ssl_prefer_server_ciphers on;

    ##
    # Logging Settings
    ##

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log;

    ##
    # Gzip Settings
    ##

    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    ##
    # Virtual Host Configs
    ##

    include /etc/nginx/conf.d/*.conf;
    include /etc/nginx/sites-enabled/*;
}
`
)
