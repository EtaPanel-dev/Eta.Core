package models

type ProcessInfo struct {
	PID     int     `json:"pid"`
	Name    string  `json:"name"`
	State   string  `json:"state"`
	CPU     float64 `json:"cpu"`
	Memory  uint64  `json:"memory"`
	Command string  `json:"command"`
}
type InterfaceStats struct {
	RxBytes   uint64
	TxBytes   uint64
	RxPackets uint64
	TxPackets uint64
}
type SystemInfo struct {
	CPU     CPUInfo     `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	Disk    []DiskInfo  `json:"disk"`
	Network NetworkInfo `json:"network"`
	System  OSInfo      `json:"system"`
}

type CPUInfo struct {
	Cores       int       `json:"cores"`
	Model       string    `json:"model"`
	Usage       float64   `json:"usage"`
	LoadAverage []float64 `json:"loadAverage"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
	Buffers     uint64  `json:"buffers"`
	Cached      uint64  `json:"cached"`
	SwapTotal   uint64  `json:"swapTotal"`
	SwapUsed    uint64  `json:"swapUsed"`
}

type DiskInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"usedPercent"`
}

type NetworkInfo struct {
	Interfaces  []NetworkInterface  `json:"interfaces"`
	Connections []NetworkConnection `json:"connections"`
}

type NetworkInterface struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	MTU       int      `json:"mtu"`
	Flags     []string `json:"flags"`
	RxBytes   uint64   `json:"rxBytes"`
	TxBytes   uint64   `json:"txBytes"`
	RxPackets uint64   `json:"rxPackets"`
	TxPackets uint64   `json:"txPackets"`
}

type NetworkConnection struct {
	Protocol   string `json:"protocol"`
	LocalAddr  string `json:"localAddr"`
	RemoteAddr string `json:"remoteAddr"`
	State      string `json:"state"`
	PID        int    `json:"pid"`
}

type OSInfo struct {
	Hostname  string `json:"hostname"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Kernel    string `json:"kernel"`
	Uptime    int64  `json:"uptime"`
	Processes int    `json:"processes"`
}
