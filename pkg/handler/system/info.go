package system

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models"
	"github.com/gin-gonic/gin"
)

// GetSystemInfo 获取系统信息
// @Summary 获取系统信息
// @Description 获取系统的CPU、内存、磁盘、网络等综合信息
// @Tags 系统监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=models.SystemInfo} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/system [get]
func GetSystemInfo(c *gin.Context) {
	info := models.SystemInfo{
		Cpu:     getCPUInfo(),
		Memory:  getMemoryInfo(),
		Disk:    getDiskInfo(),
		Network: getNetworkInfo(),
		System:  getOSInfo(),
	}

	handler.Respond(c, http.StatusOK, nil, info)
}

// getCPUInfo 获取CPU信息
func getCPUInfo() models.CpuInfo {
	cpuInfo := models.CpuInfo{
		Cores: runtime.NumCPU(),
	}

	if runtime.GOOS == "darwin" {
		// Darwin 系统的 CPU 信息获取
		cpuInfo.Model = getDarwinCPUModel()
		cpuInfo.LoadAverage = getDarwinLoadAverage()
		cpuInfo.Usage = getDarwinCPUUsage()
	} else {
		// Linux 系统的 CPU 信息获取
		// 读取CPU型号
		if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
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
		if data, err := os.ReadFile("/proc/loadavg"); err == nil {
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
	}

	return cpuInfo
}

// getDarwinCPUModel 获取 Darwin 系统的 CPU 型号
func getDarwinCPUModel() string {
	cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown"
	}
	return strings.TrimSpace(string(output))
}

// getDarwinLoadAverage 获取 Darwin 系统的负载平均值
func getDarwinLoadAverage() []float64 {
	cmd := exec.Command("sysctl", "-n", "vm.loadavg")
	output, err := cmd.Output()
	if err != nil {
		return []float64{}
	}

	// 输出格式: { 1.50 1.25 1.10 }
	line := strings.TrimSpace(string(output))
	line = strings.Trim(line, "{}")
	fields := strings.Fields(line)

	var loadAvg []float64
	for i := 0; i < 3 && i < len(fields); i++ {
		if load, err := strconv.ParseFloat(fields[i], 64); err == nil {
			loadAvg = append(loadAvg, load)
		}
	}

	return loadAvg
}

// getDarwinCPUUsage 获取 Darwin 系统的 CPU 使用率
func getDarwinCPUUsage() float64 {
	// 使用 top 命令获取 CPU 使用率
	cmd := exec.Command("top", "-l", "1", "-n", "0")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing top command:", err)
		return 0
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "CPU usage:") {
			parts := strings.Split(line, ",")
			var userUsage, sysUsage float64

			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.Contains(part, "% user") {
					userStr := strings.Fields(part)[0]
					userStr = strings.TrimSuffix(userStr, "%")
					if val, err := strconv.ParseFloat(userStr, 64); err == nil {
						userUsage = val
					} else {
						fmt.Println("Error parsing user CPU usage:", err)
					}
				} else if strings.Contains(part, "% sys") {
					sysStr := strings.Fields(part)[0]
					sysStr = strings.TrimSuffix(sysStr, "%")
					if val, err := strconv.ParseFloat(sysStr, 64); err == nil {
						sysUsage = val
					} else {
						fmt.Println("Error parsing system CPU usage:", err)
					}
				}
			}

			totalUsage := userUsage + sysUsage
			if totalUsage > 100 {
				totalUsage = 100
			}
			fmt.Printf("Parsed CPU Usage: User=%.2f, Sys=%.2f, Total=%.2f\n", userUsage, sysUsage, totalUsage)
			return totalUsage
		}
	}
	fmt.Println("CPU usage line not found in top output.")
	return 0
}

// getMemoryInfo 获取内存信息
func getMemoryInfo() models.MemoryInfo {
	if runtime.GOOS == "darwin" {
		return getDarwinMemoryInfo()
	}

	// Linux 系统的内存信息获取
	memInfo := models.MemoryInfo{}

	data, err := os.ReadFile("/proc/meminfo")
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

	// 内存信息
	memInfo.Total = memMap["MemTotal"]
	memInfo.Available = memMap["MemAvailable"]
	memInfo.Buffers = memMap["Buffers"]
	memInfo.Cached = memMap["Cached"]

	// Swap 信息
	memInfo.SwapTotal = memMap["SwapTotal"]
	swapFree := memMap["SwapFree"]
	memInfo.SwapUsed = memInfo.SwapTotal - swapFree

	// 修复内存使用量计算
	if memInfo.Total > 0 {
		if memInfo.Available > 0 {
			// MemAvailable 是更准确的可用内存计算
			memInfo.Used = memInfo.Total - memInfo.Available
		} else {
			// 如果MemAvailable不可用，使用传统计算方法
			free := memMap["MemFree"]
			sReclaimable := memMap["SReclaimable"]
			memInfo.Used = memInfo.Total - free - memInfo.Buffers - memInfo.Cached - sReclaimable
		}

		// 确保Used不会为负数或超过总内存
		if memInfo.Used > memInfo.Total {
			memInfo.Used = memInfo.Total
		}
		if memInfo.Used < 0 {
			memInfo.Used = 0
		}

		memInfo.UsedPercent = float64(memInfo.Used) / float64(memInfo.Total) * 100
	}

	return memInfo
}

// getDarwinMemoryInfo 获取 Darwin 系统的内存信息
func getDarwinMemoryInfo() models.MemoryInfo {
	memInfo := models.MemoryInfo{}

	// 获取总内存
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	if output, err := cmd.Output(); err == nil {
		if total, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64); err == nil {
			memInfo.Total = total
		}
	}

	// 使用 vm_stat 命令获取内存使用情况
	cmd = exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing vm_stat command:", err)
		return memInfo
	}

	// 获取页面大小
	pageSize := uint64(4096) // 默认页面大小
	cmd = exec.Command("sysctl", "-n", "vm.pagesize")
	if psOutput, err := cmd.Output(); err == nil {
		if ps, err := strconv.ParseUint(strings.TrimSpace(string(psOutput)), 10, 64); err == nil {
			pageSize = ps
		}
	}

	lines := strings.Split(string(output), "\n")
	var freePages, inactivePages, speculativePages uint64

	for _, line := range lines {
		if strings.Contains(line, "Pages free:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				if val, err := strconv.ParseUint(strings.TrimSuffix(parts[2], "."), 10, 64); err == nil {
					freePages = val
				}
			}
		} else if strings.Contains(line, "Pages inactive:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				if val, err := strconv.ParseUint(strings.TrimSuffix(parts[2], "."), 10, 64); err == nil {
					inactivePages = val
				}
			}
		} else if strings.Contains(line, "Pages speculative:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				if val, err := strconv.ParseUint(strings.TrimSuffix(parts[2], "."), 10, 64); err == nil {
					speculativePages = val
				}
			}
		}
	}

	// 计算内存使用情况
	// Available memory = free + speculative pages (can be reclaimed immediately)
	memInfo.Available = (freePages + speculativePages) * pageSize

	// Used memory = total - available (最准确的计算方法)
	memInfo.Used = memInfo.Total - memInfo.Available

	// Cached memory (inactive pages are often considered cached)
	memInfo.Cached = inactivePages * pageSize

	// 验证计算结果的合理性
	if memInfo.Total > 0 {
		// 确保 Used 不为负数
		if memInfo.Used < 0 {
			memInfo.Used = 0
		}

		// 确保 Used 不超过 Total
		if memInfo.Used > memInfo.Total {
			memInfo.Used = memInfo.Total
		}

		memInfo.UsedPercent = float64(memInfo.Used) / float64(memInfo.Total) * 100
	}

	// 获取交换空间信息
	getDarwinSwapInfo(&memInfo)

	return memInfo
}

// getDarwinSwapInfo 获取 Darwin 系统的交换空间信息
func getDarwinSwapInfo(memInfo *models.MemoryInfo) {
	cmd := exec.Command("sysctl", "-n", "vm.swapusage")
	output, err := cmd.Output()
	if err != nil {
		// 如果获取不到swap信息，设置为0
		memInfo.SwapTotal = 0
		memInfo.SwapUsed = 0
		return
	}

	swapInfo := strings.TrimSpace(string(output))
	total, used := parseDarwinSwapUsage(swapInfo)

	memInfo.SwapTotal = total
	memInfo.SwapUsed = used
}

// parseDarwinSwapUsage 解析 Darwin 系统的交换空间使用信息
func parseDarwinSwapUsage(swapInfo string) (total, used uint64) {
	// 输出格式: total = 2048.00M  used = 0.00M  free = 2048.00M  (encrypted)
	parts := strings.Fields(swapInfo)

	for i, part := range parts {
		if strings.Contains(part, "total") && i+2 < len(parts) {
			if totalSize := parseSwapSize(parts[i+2]); totalSize > 0 {
				total = totalSize
			}
		} else if strings.Contains(part, "used") && i+2 < len(parts) {
			if usedSize := parseSwapSize(parts[i+2]); usedSize > 0 {
				used = usedSize
			}
		}
	}

	return total, used
}

// getDiskInfo 获取磁盘信息
func getDiskInfo() []models.DiskInfo {
	if runtime.GOOS == "darwin" {
		return getDarwinDiskInfo()
	}

	// Linux 系统的磁盘信息获取
	var diskInfos []models.DiskInfo

	// 读取挂载信息
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return diskInfos
	}

	lines := strings.Split(string(data), "\n")
	seenMountpoints := make(map[string]bool)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			device := fields[0]
			mountpoint := fields[1]
			fstype := fields[2]

			// 避免重复的挂载点
			if seenMountpoints[mountpoint] {
				continue
			}

			// 只处理真实的磁盘设备和常见文件系统
			if (strings.HasPrefix(device, "/dev/") &&
				!strings.Contains(device, "loop") &&
				!strings.Contains(device, "ram")) ||
				fstype == "tmpfs" {

				// 排除系统挂载点
				if strings.HasPrefix(mountpoint, "/proc") ||
					strings.HasPrefix(mountpoint, "/sys") ||
					strings.HasPrefix(mountpoint, "/dev") ||
					strings.HasPrefix(mountpoint, "/run") ||
					mountpoint == "/boot/efi" {
					continue
				}

				var stat syscall.Statfs_t
				if err := syscall.Statfs(mountpoint, &stat); err == nil {
					total := uint64(stat.Blocks) * uint64(stat.Bsize)
					available := uint64(stat.Bavail) * uint64(stat.Bsize)

					// 跳过很小的文件系统（小于1MB）
					if total < 1024*1024 {
						continue
					}

					used := total - available
					usedPercent := float64(0)
					if total > 0 {
						usedPercent = float64(used) / float64(total) * 100
					}

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
					seenMountpoints[mountpoint] = true
				}
			}
		}
	}

	return diskInfos
}

// getDarwinDiskInfo 获取 Darwin 系统的磁盘信息
func getDarwinDiskInfo() []models.DiskInfo {
	var diskInfos []models.DiskInfo

	// 获取挂载点信息
	cmd := exec.Command("mount")
	output, err := cmd.Output()
	if err != nil {
		return diskInfos
	}

	lines := strings.Split(string(output), "\n")
	seenMountpoints := make(map[string]bool)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		device := fields[0]
		mountpoint := fields[2] // Mount point is usually the 3rd field

		// Skip special or irrelevant mount points
		if strings.HasPrefix(device, "map") ||
			strings.HasPrefix(device, "devfs") ||
			strings.HasPrefix(device, "procfs") ||
			strings.HasPrefix(device, "tmpfs") ||
			strings.HasPrefix(device, "autofs") ||
			strings.HasPrefix(mountpoint, "/System/Volumes/Preboot") ||
			strings.HasPrefix(mountpoint, "/System/Volumes/Update") ||
			strings.HasPrefix(mountpoint, "/private/var/vm") || // Swap space, not disk
			strings.HasPrefix(mountpoint, "/dev") || // /dev itself
			strings.HasPrefix(mountpoint, "/Volumes/Recovery") { // Recovery partitions
			continue
		}

		// Ensure we only process each mountpoint once
		if seenMountpoints[mountpoint] {
			continue
		}
		seenMountpoints[mountpoint] = true

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mountpoint, &stat); err == nil {
			total := uint64(stat.Blocks) * uint64(stat.Bsize)
			available := uint64(stat.Bavail) * uint64(stat.Bsize)
			used := total - available

			// Skip very small filesystems (less than 1MB)
			if total < 1024*1024 {
				continue
			}

			usedPercent := float64(0)
			if total > 0 {
				usedPercent = float64(used) / float64(total) * 100
			}

			diskInfo := models.DiskInfo{
				Device:      device,
				MountPoint:  mountpoint,
				FsType:      "", // mount command doesn't directly give fstype in a simple field
				Total:       total,
				Used:        used,
				Available:   available,
				UsedPercent: usedPercent,
			}

			// Try to get FsType from mount output if available (e.g., "type apfs")
			for _, field := range fields {
				if strings.HasPrefix(field, "type") {
					diskInfo.FsType = strings.TrimPrefix(field, "type ")
					break
				}
			}
			// Fallback for FsType if not found
			if diskInfo.FsType == "" && len(fields) > 4 {
				// Sometimes fstype is in parentheses like (apfs, local, journaled)
				if strings.HasPrefix(fields[3], "(") && strings.HasSuffix(fields[3], ",") {
					diskInfo.FsType = strings.TrimSuffix(strings.TrimPrefix(fields[3], "("), ",")
				}
			}

			diskInfos = append(diskInfos, diskInfo)
		}
	}

	return diskInfos
}

// getNetworkInfo 获取网络信息
func getNetworkInfo() models.NetworkInfo {
	networkInfo := models.NetworkInfo{
		Interfaces: getNetworkInterfaces(),
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
	if runtime.GOOS == "darwin" {
		return getDarwinInterfaceStats(ifaceName)
	}

	// Linux 系统的网络接口统计信息获取
	stats := models.InterfaceStats{}

	data, err := os.ReadFile("/proc/net/dev")
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

// getDarwinInterfaceStats 获取 Darwin 系统的网络接口统计信息
func getDarwinInterfaceStats(ifaceName string) models.InterfaceStats {
	stats := models.InterfaceStats{}

	cmd := exec.Command("netstat", "-ibn")
	output, err := cmd.Output()
	if err != nil {
		return stats
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ifaceName) {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				if rxBytes, err := strconv.ParseUint(fields[6], 10, 64); err == nil {
					stats.RxBytes = rxBytes
				}
				if rxPackets, err := strconv.ParseUint(fields[4], 10, 64); err == nil {
					stats.RxPackets = rxPackets
				}
				if txBytes, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
					stats.TxBytes = txBytes
				}
				if txPackets, err := strconv.ParseUint(fields[7], 10, 64); err == nil {
					stats.TxPackets = txPackets
				}
				break
			}
		}
	}

	return stats
}

// getNetworkConnections 获取网络连接信息
func getNetworkConnections() []models.NetworkConnection {
	if runtime.GOOS == "darwin" {
		return getDarwinNetworkConnections()
	}

	// Linux 系统的网络连接信息获取
	var connections []models.NetworkConnection

	// TCP连接
	tcpConns := parseNetworkConnections("/proc/net/tcp", "tcp")
	connections = append(connections, tcpConns...)

	// UDP连接
	udpConns := parseNetworkConnections("/proc/net/udp", "udp")
	connections = append(connections, udpConns...)

	return connections
}

// getDarwinNetworkConnections 获取 Darwin 系统的网络连接信息
func getDarwinNetworkConnections() []models.NetworkConnection {
	var connections []models.NetworkConnection

	cmd := exec.Command("netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		return connections
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 6 && (fields[0] == "tcp4" || fields[0] == "tcp6" || fields[0] == "udp4" || fields[0] == "udp6") {
			protocol := strings.TrimRight(fields[0], "46")
			localAddr := fields[3]
			remoteAddr := fields[4]
			state := ""
			if len(fields) >= 6 {
				state = fields[5]
			}

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

// parseNetworkConnections 解析网络连接文件 (Linux)
func parseNetworkConnections(filename, protocol string) []models.NetworkConnection {
	var connections []models.NetworkConnection

	data, err := os.ReadFile(filename)
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

// parseAddress 解析地址 (Linux)
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

// parseConnectionState 解析连接状态 (Linux)
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

	if runtime.GOOS == "darwin" {
		// Darwin 系统的信息获取
		// 获取内核版本
		cmd := exec.Command("uname", "-r")
		if output, err := cmd.Output(); err == nil {
			osInfo.Kernel = strings.TrimSpace(string(output))
		}

		// 获取系统运行时间
		osInfo.Uptime = getDarwinUptime()

		// 获取进程数量
		osInfo.Processes = getDarwinProcessCount()
	} else {
		// Linux 系统的信息获取
		// 获取内核版本
		if data, err := os.ReadFile("/proc/version"); err == nil {
			version := strings.TrimSpace(string(data))
			fields := strings.Fields(version)
			if len(fields) >= 3 {
				osInfo.Kernel = fields[2]
			}
		}

		// 获取系统运行时间
		if data, err := os.ReadFile("/proc/uptime"); err == nil {
			fields := strings.Fields(string(data))
			if len(fields) >= 1 {
				if uptime, err := strconv.ParseFloat(fields[0], 64); err == nil {
					osInfo.Uptime = int64(uptime)
				}
			}
		}

		// 获取进程数量
		osInfo.Processes = getProcessCount()
	}

	return osInfo
}

// getDarwinUptime 获取 Darwin 系统运行时间
func getDarwinUptime() int64 {
	cmd := exec.Command("sysctl", "-n", "kern.boottime")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	// 输出格式: { sec = 1234567890, usec = 123456 } Mon Jan  1 00:00:00 2024
	line := strings.TrimSpace(string(output))
	if strings.Contains(line, "sec = ") {
		start := strings.Index(line, "sec = ") + 6
		end := strings.Index(line[start:], ",")
		if end != -1 {
			bootTimeStr := line[start : start+end]
			if bootTime, err := strconv.ParseInt(bootTimeStr, 10, 64); err == nil {
				return time.Now().Unix() - bootTime
			}
		}
	}

	return 0
}

// getDarwinProcessCount 获取 Darwin 系统进程数量
func getDarwinProcessCount() int {
	cmd := exec.Command("ps", "ax")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(output), "\n")
	// 减去标题行和空行
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	// 减去标题行
	if count > 0 {
		count--
	}

	return count
}

// GetCPUInfo 单独获取CPU信息
// @Summary 获取CPU信息
// @Description 获取CPU核心数、型号、使用率和负载平均值
// @Tags 系统监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=models.CpuInfo} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/system/cpu [get]
func GetCPUInfo(c *gin.Context) {
	handler.Respond(c, http.StatusOK, nil, getCPUInfo())
}

// GetMemoryInfo 单独获取内存信息
// @Summary 获取内存信息
// @Description 获取系统内存使用情况，包括总内存、已用内存、缓存等
// @Tags 系统监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=models.MemoryInfo} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/system/memory [get]
func GetMemoryInfo(c *gin.Context) {
	handler.Respond(c, http.StatusOK, nil, getMemoryInfo())
}

// GetDiskInfo 单独获取磁盘信息
// @Summary 获取磁盘信息
// @Description 获取所有磁盘分区的使用情况
// @Tags 系统监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=[]models.DiskInfo} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/system/disk [get]
func GetDiskInfo(c *gin.Context) {
	handler.Respond(c, http.StatusOK, nil, getDiskInfo())
}

// GetNetworkInfo 单独获取网络信息
// @Summary 获取网络信息
// @Description 获取网络接口和连接信息
// @Tags 系统监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=models.NetworkInfo} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/system/network [get]
func GetNetworkInfo(c *gin.Context) {
	handler.Respond(c, http.StatusOK, nil, getNetworkInfo())
}

// getCPUUsage 计算CPU使用率 (Linux)
func getCPUUsage() float64 {
	// 读取两次/proc/stat来计算CPU使用率
	stat1 := readCPUStat()
	if len(stat1) < 4 {
		return 0
	}

	time.Sleep(100 * time.Millisecond)
	stat2 := readCPUStat()
	if len(stat2) < 4 {
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

	usage := float64(totalDiff-idleDiff) / float64(totalDiff) * 100
	if usage < 0 {
		return 0
	}
	if usage > 100 {
		return 100
	}

	return usage
}

// readCPUStat 读取CPU统计信息 (Linux)
func readCPUStat() []uint64 {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return nil
	}

	// 第一行是总CPU统计信息
	line := lines[0]
	if !strings.HasPrefix(line, "cpu ") {
		return nil
	}

	fields := strings.Fields(line)
	if len(fields) < 8 {
		return nil
	}

	// 解析CPU时间统计
	// fields[0] = "cpu"
	// fields[1] = user
	// fields[2] = nice
	// fields[3] = system
	// fields[4] = idle
	// fields[5] = iowait
	// fields[6] = irq
	// fields[7] = softirq
	// fields[8] = steal (可选)
	// fields[9] = guest (可选)
	// fields[10] = guest_nice (可选)

	var stats []uint64
	for i := 1; i < len(fields) && i <= 10; i++ {
		if val, err := strconv.ParseUint(fields[i], 10, 64); err == nil {
			stats = append(stats, val)
		} else {
			break
		}
	}

	return stats
}

// getSysctlUint64 辅助函数，用于从 sysctl 获取 uint64 值
func getSysctlUint64(name string) uint64 {
	cmd := exec.Command("sysctl", "-n", name)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	val, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// parseSwapSize 解析交换空间大小（支持 M, G 等单位）
func parseSwapSize(sizeStr string) uint64 {
	sizeStr = strings.TrimSpace(sizeStr)
	if len(sizeStr) == 0 {
		return 0
	}

	unit := sizeStr[len(sizeStr)-1:]
	numStr := sizeStr[:len(sizeStr)-1]

	size, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	switch strings.ToUpper(unit) {
	case "K":
		return uint64(size * 1024)
	case "M":
		return uint64(size * 1024 * 1024)
	case "G":
		return uint64(size * 1024 * 1024 * 1024)
	default:
		// 假设是字节
		return uint64(size)
	}
}
