package models

import (
	"gorm.io/gorm"
	"time"
)

type Server struct {
	gorm.Model
	Id          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	IP          string    `json:"ip" gorm:"not null"`
	Username    string    `json:"username"`
	Status      string    `json:"status" gorm:"default:offline"`
	OS          string    `json:"os"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SystemMetric struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ServerID   uint      `json:"server_id"`
	CPUUsage   float64   `json:"cpu_usage"`
	MemUsage   float64   `json:"mem_usage"`
	MemMax     float64   `json:"mem_max"`
	DiskUsage  float64   `json:"disk_usage"`
	DiskMax    float64   `json:"disk_max"`
	NetworkIn  uint64    `json:"network_in"`
	NetworkOut uint64    `json:"network_out"`
	Timestamp  time.Time `json:"timestamp"`
	Server     Server    `json:"server" gorm:"foreignKey:ServerID"`
}

//
