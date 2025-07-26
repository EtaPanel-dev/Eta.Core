package models

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

// DockerImage Docker镜像信息
type DockerImage struct {
	ID         string    `json:"id"`
	Repository string    `json:"repository"`
	Tag        string    `json:"tag"`
	Created    time.Time `json:"created"`
	Size       string    `json:"size"`
}

// DockerContainer Docker容器信息
type DockerContainer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	Ports   string `json:"ports"`
	Created string `json:"created"`
}

// DockerRegistry Docker镜像源配置
type DockerRegistry struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Default  bool   `json:"default"`
}

// Upgrader WebSocket升级器
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}
