package system

import (
	"fmt"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/handler"
	"github.com/EtaPanel-dev/Eta-Panel/core/pkg/models"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// GetSystemInfo 获取系统信息
func GetSystemInfo(c *gin.Context) {
	info := models.SystemInfo{
		Cpu:     getCPUInfo(),
		Memory:  getMemoryInfo(),
		Disk:    getDiskInfo(),
		Network: getNetworkInfo(),
		System:  getOSInfo(),
	}

	handler.Respond(c, http.StatusOK, info, nil)
}

// getCPUInfo 获取CPU信息
func getCPUInfo() models.CpuInfo {
	cpuInfo := models.CpuInfo{
		Cores: runtime.NumCPU(),
	}

	// 读取CPU型号
	if data, err := ioutil.ReadFile("/proc/cpuinfo"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "model name") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					cpuInfo.Model = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	// 读取负载平均值
	if data, err := ioutil.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 3 {
			for i := 0; i < 3; i++ {
				if load, err := strconv.ParseFloat(fields[i], 64); err == nil {
					cpuInfo.LoadAverage = append(cpuInfo.LoadAverage, load)
				}
			}
		}
	}

	// 计算CPU使用率
	cpuInfo.Usage = getCPUUsage()

	return cpuInfo
}

// getCPUUsage 计算CPU使用率
func getCPUUsage() float64 {
	// 读取两次/proc/stat来计算CPU使用率
	stat1 := readCPUStat()
	time.Sleep(100 * time.Millisecond)
	stat2 := readCPUStat()

	if len(stat1) < 4 || len(stat2) < 4 {
		return 0
	}

	idle1 := stat1[3]
	idle2 := stat2[3]

	total1 := uint64(0)
	total2 := uint64(0)

	for _, val := range stat1 {
		total1 += val
	}
	for _, val := range stat2 {
		total2 += val
	}

	totalDiff := total2 - total1
	idleDiff := idle2 - idle1

	if totalDiff == 0 {
		return 0
	}

	return float64(totalDiff-idleDiff) / float64(totalDiff) * 100
}

// readCPUStat 读取CPU统计信息
func readCPUStat() []uint64 {
	data, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return nil
	}

	fields := strings.Fields(lines[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return nil
	}

	var stats []uint64
	for i := 1; i < len(fields); i++ {
		if val, err := strconv.ParseUint(fields[i], 10, 64); err == nil {
			stats = append(stats, val)
		}
	}

	return stats
}

// getMemoryInfo 获取内存信息
func getMemoryInfo() models.MemoryInfo {
	memInfo := models.MemoryInfo{}

	data, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return memInfo
	}

	lines := strings.Split(string(data), "\n")
	memMap := make(map[string]uint64)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			key := strings.TrimSuffix(fields[0], ":")
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memMap[key] = val * 1024 // 转换为字节
			}
		}
	}

	memInfo.Total = memMap["MemTotal"]
	memInfo.Available = memMap["MemAvailable"]
	memInfo.Buffers = memMap["Buffers"]
	memInfo.Cached = memMap["Cached"]
	memInfo.SwapTotal = memMap["SwapTotal"]
	memInfo.SwapUsed = memMap["SwapTotal"] - memMap["SwapFree"]

	if memInfo.Available > 0 {
		memInfo.Used = memInfo.Total - memInfo.Available
		memInfo.UsedPercent = float64(memInfo.Used) / float64(memInfo.Total) * 100
	}

	return memInfo
}

// getDiskInfo 获取磁盘信息
func getDiskInfo() []models.DiskInfo {
	var diskInfos []models.DiskInfo

	// 读取挂载信息
	data, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return diskInfos
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			device := fields[0]
			mountpoint := fields[1]
			fstype := fields[2]

			// 只处理真实的磁盘设备
			if strings.HasPrefix(device, "/dev/") &&
				!strings.Contains(device, "loop") &&
				!strings.Contains(mountpoint, "/proc") &&
				!strings.Contains(mountpoint, "/sys") &&
				!strings.Contains(mountpoint, "/dev") {

				var stat syscall.Statfs_t
				if err := syscall.Statfs(mountpoint, &stat); err == nil {
					total := uint64(stat.Blocks) * uint64(stat.Bsize)
					available := uint64(stat.Bavail) * uint64(stat.Bsize)
					used := total - available
					usedPercent := float64(used) / float64(total) * 100

					diskInfo := models.DiskInfo{
						Device:      device,
						MountPoint:  mountpoint,
						FsType:      fstype,
						Total:       total,
						Used:        used,
						Available:   available,
						UsedPercent: usedPercent,
					}

					diskInfos = append(diskInfos, diskInfo)
				}
			}
		}
	}

	return diskInfos
}

// getNetworkInfo 获取网络信息
func getNetworkInfo() models.NetworkInfo {
	networkInfo := models.NetworkInfo{
		Interfaces:  getNetworkInterfaces(),
		Connections: getNetworkConnections(),
	}

	return networkInfo
}

// getNetworkInterfaces 获取网络接口信息
func getNetworkInterfaces() []models.NetworkInterface {
	var interfaces []models.NetworkInterface

	ifaces, err := net.Interfaces()
	if err != nil {
		return interfaces
	}

	for _, iface := range ifaces {
		netInterface := models.NetworkInterface{
			Name: iface.Name,
			Mtu:  iface.MTU,
		}

		// 获取接口标志
		flags := []string{}
		if iface.Flags&net.FlagUp != 0 {
			flags = append(flags, "UP")
		}
		if iface.Flags&net.FlagLoopback != 0 {
			flags = append(flags, "LOOPBACK")
		}
		if iface.Flags&net.FlagMulticast != 0 {
			flags = append(flags, "MULTICAST")
		}
		netInterface.Flags = flags

		// 获取IP地址
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				netInterface.Addresses = append(netInterface.Addresses, addr.String())
			}
		}

		// 获取网络统计信息
		stats := getInterfaceStats(iface.Name)
		netInterface.RxBytes = stats.RxBytes
		netInterface.TxBytes = stats.TxBytes
		netInterface.RxPackets = stats.RxPackets
		netInterface.TxPackets = stats.TxPackets

		interfaces = append(interfaces, netInterface)
	}

	return interfaces
}

// getInterfaceStats 获取网络接口统计信息
func getInterfaceStats(ifaceName string) models.InterfaceStats {
	stats := models.InterfaceStats{}

	data, err := ioutil.ReadFile("/proc/net/dev")
	if err != nil {
		return stats
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, ifaceName+":") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				if rxBytes, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					stats.RxBytes = rxBytes
				}
				if rxPackets, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
					stats.RxPackets = rxPackets
				}
				if txBytes, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
					stats.TxBytes = txBytes
				}
				if txPackets, err := strconv.ParseUint(fields[10], 10, 64); err == nil {
					stats.TxPackets = txPackets
				}
			}
			break
		}
	}

	return stats
}

// getNetworkConnections 获取网络连接信息
func getNetworkConnections() []models.NetworkConnection {
	var connections []models.NetworkConnection

	// TCP连接
	tcpConns := parseNetworkConnections("/proc/net/tcp", "tcp")
	connections = append(connections, tcpConns...)

	// UDP连接
	udpConns := parseNetworkConnections("/proc/net/udp", "udp")
	connections = append(connections, udpConns...)

	return connections
}

// parseNetworkConnections 解析网络连接文件
func parseNetworkConnections(filename, protocol string) []models.NetworkConnection {
	var connections []models.NetworkConnection

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return connections
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i == 0 { // 跳过标题行
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 10 {
			localAddr := parseAddress(fields[1])
			remoteAddr := parseAddress(fields[2])
			state := parseConnectionState(fields[3])

			connection := models.NetworkConnection{
				Protocol:   protocol,
				LocalAddr:  localAddr,
				RemoteAddr: remoteAddr,
				State:      state,
			}

			connections = append(connections, connection)
		}
	}

	return connections
}

// parseAddress 解析地址
func parseAddress(addr string) string {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return addr
	}

	// 解析IP地址
	ipHex := parts[0]
	portHex := parts[1]

	if len(ipHex) == 8 { // IPv4
		ip := make([]byte, 4)
		for i := 0; i < 4; i++ {
			if val, err := strconv.ParseUint(ipHex[i*2:(i+1)*2], 16, 8); err == nil {
				ip[3-i] = byte(val) // 小端序
			}
		}

		if port, err := strconv.ParseUint(portHex, 16, 16); err == nil {
			return fmt.Sprintf("%d.%d.%d.%d:%d", ip[0], ip[1], ip[2], ip[3], port)
		}
	}

	return addr
}

// parseConnectionState 解析连接状态
func parseConnectionState(state string) string {
	states := map[string]string{
		"01": "ESTABLISHED",
		"02": "SYN_SENT",
		"03": "SYN_RECV",
		"04": "FIN_WAIT1",
		"05": "FIN_WAIT2",
		"06": "TIME_WAIT",
		"07": "CLOSE",
		"08": "CLOSE_WAIT",
		"09": "LAST_ACK",
		"0A": "LISTEN",
		"0B": "CLOSING",
	}

	if stateName, exists := states[state]; exists {
		return stateName
	}

	return "UNKNOWN"
}

// getOSInfo 获取操作系统信息
func getOSInfo() models.OSInfo {
	osInfo := models.OSInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// 获取主机名
	if hostname, err := os.Hostname(); err == nil {
		osInfo.Hostname = hostname
	}

	// 获取内核版本
	if data, err := ioutil.ReadFile("/proc/version"); err == nil {
		version := strings.TrimSpace(string(data))
		fields := strings.Fields(version)
		if len(fields) >= 3 {
			osInfo.Kernel = fields[2]
		}
	}

	// 获取系统运行时间
	if data, err := ioutil.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 1 {
			if uptime, err := strconv.ParseFloat(fields[0], 64); err == nil {
				osInfo.Uptime = int64(uptime)
			}
		}
	}

	// 获取进程数量
	osInfo.Processes = getProcessCount()

	return osInfo
}

// GetCPUInfo 单独获取CPU信息
func GetCPUInfo(c *gin.Context) {
	handler.Respond(c, http.StatusOK, nil, getCPUInfo())
}

// GetMemoryInfo 单独获取内存信息
func GetMemoryInfo(c *gin.Context) {
	handler.Respond(c, http.StatusOK, nil, getMemoryInfo())
}

// GetDiskInfo 单独获取磁盘信息
func GetDiskInfo(c *gin.Context) {
	handler.Respond(c, http.StatusOK, nil, getDiskInfo())
}

// GetNetworkInfo 单独获取网络信息
func GetNetworkInfo(c *gin.Context) {
	handler.Respond(c, http.StatusOK, nil, getNetworkInfo())
}
