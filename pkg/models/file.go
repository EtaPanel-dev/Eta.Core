package models

import "time"

type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	Mode        string    `json:"mode"`
	ModTime     time.Time `json:"modTime"`
	IsDir       bool      `json:"isDir"`
	IsSymlink   bool      `json:"isSymlink"`
	Permissions string    `json:"permissions"`
	Owner       string    `json:"owner"`
	Group       string    `json:"group"`
}

// ProtectedDirs 受保护的目录列表
var ProtectedDirs = []string{
	"/etc/passwd",
	"/etc/shadow",
	"/etc/sudoers",
	"/boot",
	"/sys",
	"/proc",
	"/dev",
}
