package checks

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
)

type SystemInfo struct {
	Hostname     string
	OS           string
	OSVersion    string
	Kernel       string
	Architecture string
	CPUModel     string
	CPUCores     int
	Uptime       string
	LoadAverage  []float64
}

func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	info.Hostname, _ = os.Hostname()
	info.Architecture = runtime.GOARCH

	hostInfo, err := host.Info()
	if err == nil {
		info.OS = hostInfo.Platform
		info.OSVersion = hostInfo.PlatformVersion
		info.Kernel = hostInfo.KernelVersion
		info.Uptime = formatUptime(hostInfo.Uptime)
	}

	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		info.CPUModel = cpuInfo[0].ModelName
		info.CPUCores = int(cpuInfo[0].Cores)
	}

	if loadAvg, err := getLoadAverage(); err == nil {
		info.LoadAverage = loadAvg
	}

	return info, nil
}

func formatUptime(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func getLoadAverage() ([]float64, error) {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 {
			loads := make([]float64, 3)
			fmt.Sscanf(fields[0], "%f", &loads[0])
			fmt.Sscanf(fields[1], "%f", &loads[1])
			fmt.Sscanf(fields[2], "%f", &loads[2])
			return loads, nil
		}
	}
	return nil, fmt.Errorf("unable to parse load average")
}

func GetDistributionInfo() (string, string, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	var name, version string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "NAME=") {
			name = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
		} else if strings.HasPrefix(line, "VERSION=") {
			version = strings.Trim(strings.TrimPrefix(line, "VERSION="), "\"")
		}
	}
	return name, version, nil
}

func IsSystemdSystem() bool {
	if _, err := exec.LookPath("systemctl"); err == nil {
		return true
	}
	return false
}

// SystemInfoCheck checks basic system information
type SystemInfoCheck struct{}

func (s SystemInfoCheck) Name() string {
	return "System Information"
}

func (s SystemInfoCheck) RequiresRoot() bool {
	return false
}

func (s SystemInfoCheck) Run() CheckResult {
	result := CheckResult{
		Name:      s.Name(),
		Severity:  SeverityInfo,
		Timestamp: time.Now(),
		Details:   []string{},
	}

	// Get system info
	sysInfo, err := GetSystemInfo()
	if err != nil {
		result.Severity = SeverityError
		result.Message = "Unable to determine system information"
		return result
	}

	result.Details = append(result.Details, fmt.Sprintf("OS: %s", sysInfo.OS))
	result.Details = append(result.Details, fmt.Sprintf("Version: %s", sysInfo.OSVersion))
	result.Details = append(result.Details, fmt.Sprintf("Kernel: %s", sysInfo.Kernel))
	result.Details = append(result.Details, fmt.Sprintf("Architecture: %s", sysInfo.Architecture))
	result.Details = append(result.Details, fmt.Sprintf("Hostname: %s", sysInfo.Hostname))
	result.Details = append(result.Details, fmt.Sprintf("Uptime: %s", sysInfo.Uptime))

	// Check if it's actually Debian or Debian-based
	osInfo, _ := getOSRelease()
	isDebian := strings.Contains(strings.ToLower(sysInfo.OS), "debian") ||
		strings.Contains(strings.ToLower(osInfo["ID"]), "debian") ||
		strings.Contains(strings.ToLower(osInfo["ID_LIKE"]), "debian")
	
	if !isDebian {
		result.Severity = SeverityWarning
		result.Message = "This doesn't appear to be a Debian-based system"
	} else {
		if strings.Contains(strings.ToLower(osInfo["ID"]), "ubuntu") {
			result.Message = fmt.Sprintf("Ubuntu %s detected (Debian-based)", sysInfo.OSVersion)
		} else if strings.Contains(strings.ToLower(osInfo["ID"]), "debian") {
			result.Message = fmt.Sprintf("Debian %s detected", sysInfo.OSVersion)
		} else {
			result.Message = fmt.Sprintf("Debian-based system detected: %s %s", sysInfo.OS, sysInfo.OSVersion)
		}
	}

	return result
}

func getOSRelease() (map[string]string, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	osInfo := make(map[string]string)
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := parts[0]
			value := strings.Trim(parts[1], "\"")
			osInfo[key] = value
		}
	}

	return osInfo, scanner.Err()
}