package summary

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/debian-doctor/debian-doctor/internal/checks"
	"github.com/debian-doctor/debian-doctor/pkg/config"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// Generator creates comprehensive system reports
type Generator struct {
	config    *config.Config
	startTime time.Time
	endTime   time.Time
}

// NewGenerator creates a new summary generator
func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		config:    cfg,
		startTime: time.Now(),
	}
}

// SystemSummary holds comprehensive system information
type SystemSummary struct {
	Timestamp       time.Time
	Duration        time.Duration
	SystemInfo      SystemInfo
	ResourceStatus  ResourceStatus
	NetworkStatus   NetworkStatus
	CheckResults    checks.Results
	HealthScore     int
	Recommendations []string
	CriticalIssues  []string
	Warnings        []string
}

// SystemInfo contains basic system information
type SystemInfo struct {
	Hostname       string
	OS             string
	Kernel         string
	Architecture   string
	CPUModel       string
	CPUCores       int
	TotalMemory    uint64
	Uptime         time.Duration
	BootTime       time.Time
	Virtualization string
}

// ResourceStatus contains resource usage information
type ResourceStatus struct {
	CPUUsage       float64
	MemoryUsed     uint64
	MemoryPercent  float64
	SwapUsed       uint64
	SwapPercent    float64
	DiskUsage      []DiskInfo
	LoadAverage    [3]float64
	ProcessCount   int
}

// DiskInfo contains disk usage details
type DiskInfo struct {
	Path        string
	Device      string
	Filesystem  string
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
}

// NetworkStatus contains network information
type NetworkStatus struct {
	Interfaces []NetworkInterface
	DNSServers []string
	Gateway    string
	Hostname   string
}

// NetworkInterface contains network interface details
type NetworkInterface struct {
	Name      string
	Addresses []string
	Status    string
	MTU       int
}

// Generate creates a comprehensive system summary
func (g *Generator) Generate(results checks.Results) (*SystemSummary, error) {
	g.endTime = time.Now()
	
	summary := &SystemSummary{
		Timestamp:    g.startTime,
		Duration:     g.endTime.Sub(g.startTime),
		CheckResults: results,
	}
	
	// Gather system information
	if err := g.gatherSystemInfo(summary); err != nil {
		return nil, fmt.Errorf("failed to gather system info: %w", err)
	}
	
	// Gather resource status
	if err := g.gatherResourceStatus(summary); err != nil {
		return nil, fmt.Errorf("failed to gather resource status: %w", err)
	}
	
	// Gather network status
	if err := g.gatherNetworkStatus(summary); err != nil {
		return nil, fmt.Errorf("failed to gather network status: %w", err)
	}
	
	// Calculate health score
	g.calculateHealthScore(summary)
	
	// Generate recommendations
	g.generateRecommendations(summary)
	
	// Extract critical issues and warnings
	summary.CriticalIssues = results.GetErrors()
	summary.Warnings = results.GetWarnings()
	
	return summary, nil
}

func (g *Generator) gatherSystemInfo(summary *SystemSummary) error {
	info := SystemInfo{}
	
	// Host information
	if hostInfo, err := host.Info(); err == nil {
		info.Hostname = hostInfo.Hostname
		info.OS = fmt.Sprintf("%s %s", hostInfo.Platform, hostInfo.PlatformVersion)
		info.Kernel = hostInfo.KernelVersion
		info.Architecture = hostInfo.KernelArch
		info.Uptime = time.Duration(hostInfo.Uptime) * time.Second
		info.BootTime = time.Unix(int64(hostInfo.BootTime), 0)
		info.Virtualization = hostInfo.VirtualizationSystem
		if info.Virtualization == "" {
			info.Virtualization = "none"
		}
	}
	
	// CPU information
	if cpuInfo, err := cpu.Info(); err == nil && len(cpuInfo) > 0 {
		info.CPUModel = cpuInfo[0].ModelName
		info.CPUCores = int(cpuInfo[0].Cores)
	}
	
	// Memory information
	if memInfo, err := mem.VirtualMemory(); err == nil {
		info.TotalMemory = memInfo.Total
	}
	
	// Runtime information
	info.Architecture = runtime.GOARCH
	
	summary.SystemInfo = info
	return nil
}

func (g *Generator) gatherResourceStatus(summary *SystemSummary) error {
	status := ResourceStatus{}
	
	// CPU usage
	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
		status.CPUUsage = cpuPercent[0]
	}
	
	// Memory usage
	if memInfo, err := mem.VirtualMemory(); err == nil {
		status.MemoryUsed = memInfo.Used
		status.MemoryPercent = memInfo.UsedPercent
	}
	
	// Swap usage
	if swapInfo, err := mem.SwapMemory(); err == nil {
		status.SwapUsed = swapInfo.Used
		status.SwapPercent = swapInfo.UsedPercent
	}
	
	// Disk usage
	if partitions, err := disk.Partitions(false); err == nil {
		for _, partition := range partitions {
			if usage, err := disk.Usage(partition.Mountpoint); err == nil {
				// Skip special filesystems
				if strings.HasPrefix(partition.Mountpoint, "/sys") ||
					strings.HasPrefix(partition.Mountpoint, "/proc") ||
					strings.HasPrefix(partition.Mountpoint, "/dev") ||
					strings.HasPrefix(partition.Mountpoint, "/run") {
					continue
				}
				
				status.DiskUsage = append(status.DiskUsage, DiskInfo{
					Path:        partition.Mountpoint,
					Device:      partition.Device,
					Filesystem:  partition.Fstype,
					Total:       usage.Total,
					Used:        usage.Used,
					Free:        usage.Free,
					UsedPercent: usage.UsedPercent,
				})
			}
		}
	}
	
	// Load average (Linux/Unix only)
	if runtime.GOOS != "windows" {
		if avg, err := getLoadAverage(); err == nil {
			status.LoadAverage = avg
		}
	}
	
	summary.ResourceStatus = status
	return nil
}

func (g *Generator) gatherNetworkStatus(summary *SystemSummary) error {
	status := NetworkStatus{}
	
	// Get hostname
	status.Hostname, _ = os.Hostname()
	
	// Network interfaces
	if interfaces, err := net.Interfaces(); err == nil {
		for _, iface := range interfaces {
			if iface.Name == "lo" {
				continue // Skip loopback
			}
			
			netIface := NetworkInterface{
				Name: iface.Name,
				MTU:  iface.MTU,
			}
			
			// Get addresses
			for _, addr := range iface.Addrs {
				netIface.Addresses = append(netIface.Addresses, addr.Addr)
			}
			
			// Determine status
			if len(iface.Addrs) > 0 {
				netIface.Status = "UP"
			} else {
				netIface.Status = "DOWN"
			}
			
			status.Interfaces = append(status.Interfaces, netIface)
		}
	}
	
	// DNS servers (from /etc/resolv.conf on Unix-like systems)
	status.DNSServers = getDNSServers()
	
	summary.NetworkStatus = status
	return nil
}

func (g *Generator) calculateHealthScore(summary *SystemSummary) {
	score := 100
	
	// Deduct for critical issues
	criticalCount := len(summary.CriticalIssues)
	score -= criticalCount * 20
	
	// Deduct for warnings
	warningCount := len(summary.Warnings)
	score -= warningCount * 5
	
	// Deduct for high resource usage
	if summary.ResourceStatus.CPUUsage > 80 {
		score -= 10
	}
	if summary.ResourceStatus.MemoryPercent > 90 {
		score -= 10
	}
	if summary.ResourceStatus.SwapPercent > 50 {
		score -= 5
	}
	
	// Check disk usage
	for _, disk := range summary.ResourceStatus.DiskUsage {
		if disk.UsedPercent > 90 {
			score -= 10
		} else if disk.UsedPercent > 80 {
			score -= 5
		}
	}
	
	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	summary.HealthScore = score
}

func (g *Generator) generateRecommendations(summary *SystemSummary) {
	recommendations := []string{}
	
	// CPU recommendations
	if summary.ResourceStatus.CPUUsage > 80 {
		recommendations = append(recommendations, 
			"High CPU usage detected. Consider identifying resource-intensive processes.")
	}
	
	// Memory recommendations
	if summary.ResourceStatus.MemoryPercent > 90 {
		recommendations = append(recommendations,
			"Memory usage is critical. Consider closing unused applications or adding more RAM.")
	} else if summary.ResourceStatus.MemoryPercent > 80 {
		recommendations = append(recommendations,
			"Memory usage is high. Monitor for memory leaks.")
	}
	
	// Swap recommendations
	if summary.ResourceStatus.SwapPercent > 50 {
		recommendations = append(recommendations,
			"High swap usage indicates memory pressure. Consider adding more RAM.")
	}
	
	// Disk recommendations
	for _, disk := range summary.ResourceStatus.DiskUsage {
		if disk.UsedPercent > 90 {
			recommendations = append(recommendations,
				fmt.Sprintf("Critical disk space on %s (%.1f%% used). Clean up immediately.", 
					disk.Path, disk.UsedPercent))
		} else if disk.UsedPercent > 80 {
			recommendations = append(recommendations,
				fmt.Sprintf("Low disk space on %s (%.1f%% used). Consider cleanup.", 
					disk.Path, disk.UsedPercent))
		}
	}
	
	// System uptime recommendation
	if summary.SystemInfo.Uptime > 30*24*time.Hour {
		recommendations = append(recommendations,
			"System has been running for over 30 days. Consider scheduling a reboot for updates.")
	}
	
	// Network recommendations
	if len(summary.NetworkStatus.Interfaces) == 0 {
		recommendations = append(recommendations,
			"No active network interfaces detected.")
	}
	
	if len(summary.NetworkStatus.DNSServers) == 0 {
		recommendations = append(recommendations,
			"No DNS servers configured. Check network settings.")
	}
	
	summary.Recommendations = recommendations
}

// FormatReport generates a human-readable report
func (s *SystemSummary) FormatReport() string {
	var b strings.Builder
	
	b.WriteString("\n=====================================\n")
	b.WriteString("     COMPREHENSIVE SYSTEM REPORT    \n")
	b.WriteString("=====================================\n\n")
	
	// Timestamp and duration
	b.WriteString(fmt.Sprintf("Report Generated: %s\n", s.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("Scan Duration: %s\n", s.Duration.Round(time.Second)))
	b.WriteString("\n")
	
	// Health Score with visual indicator
	b.WriteString("SYSTEM HEALTH SCORE\n")
	b.WriteString(fmt.Sprintf("  Score: %d/100 ", s.HealthScore))
	b.WriteString(getHealthBar(s.HealthScore))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Status: %s\n", getHealthStatus(s.HealthScore)))
	b.WriteString("\n")
	
	// System Information
	b.WriteString("SYSTEM INFORMATION\n")
	b.WriteString(fmt.Sprintf("  Hostname: %s\n", s.SystemInfo.Hostname))
	b.WriteString(fmt.Sprintf("  OS: %s\n", s.SystemInfo.OS))
	b.WriteString(fmt.Sprintf("  Kernel: %s\n", s.SystemInfo.Kernel))
	b.WriteString(fmt.Sprintf("  Architecture: %s\n", s.SystemInfo.Architecture))
	b.WriteString(fmt.Sprintf("  CPU: %s (%d cores)\n", s.SystemInfo.CPUModel, s.SystemInfo.CPUCores))
	b.WriteString(fmt.Sprintf("  Memory: %.2f GB\n", float64(s.SystemInfo.TotalMemory)/(1024*1024*1024)))
	b.WriteString(fmt.Sprintf("  Uptime: %s\n", formatDuration(s.SystemInfo.Uptime)))
	b.WriteString(fmt.Sprintf("  Boot Time: %s\n", s.SystemInfo.BootTime.Format("2006-01-02 15:04:05")))
	if s.SystemInfo.Virtualization != "none" {
		b.WriteString(fmt.Sprintf("  Virtualization: %s\n", s.SystemInfo.Virtualization))
	}
	b.WriteString("\n")
	
	// Resource Usage
	b.WriteString("RESOURCE USAGE\n")
	b.WriteString(fmt.Sprintf("  CPU Usage: %.1f%%\n", s.ResourceStatus.CPUUsage))
	b.WriteString(fmt.Sprintf("  Memory: %.2f GB / %.2f GB (%.1f%%)\n",
		float64(s.ResourceStatus.MemoryUsed)/(1024*1024*1024),
		float64(s.SystemInfo.TotalMemory)/(1024*1024*1024),
		s.ResourceStatus.MemoryPercent))
	if s.ResourceStatus.SwapUsed > 0 {
		b.WriteString(fmt.Sprintf("  Swap: %.2f GB (%.1f%%)\n",
			float64(s.ResourceStatus.SwapUsed)/(1024*1024*1024),
			s.ResourceStatus.SwapPercent))
	}
	if runtime.GOOS != "windows" && s.ResourceStatus.LoadAverage[0] > 0 {
		b.WriteString(fmt.Sprintf("  Load Average: %.2f, %.2f, %.2f\n",
			s.ResourceStatus.LoadAverage[0],
			s.ResourceStatus.LoadAverage[1],
			s.ResourceStatus.LoadAverage[2]))
	}
	b.WriteString("\n")
	
	// Disk Usage
	if len(s.ResourceStatus.DiskUsage) > 0 {
		b.WriteString("DISK USAGE\n")
		for _, disk := range s.ResourceStatus.DiskUsage {
			status := "OK"
			if disk.UsedPercent > 90 {
				status = "CRITICAL"
			} else if disk.UsedPercent > 80 {
				status = "WARNING"
			}
			b.WriteString(fmt.Sprintf("  %s (%s)\n", disk.Path, disk.Filesystem))
			b.WriteString(fmt.Sprintf("    %.2f GB / %.2f GB (%.1f%%) - %s\n",
				float64(disk.Used)/(1024*1024*1024),
				float64(disk.Total)/(1024*1024*1024),
				disk.UsedPercent,
				status))
		}
		b.WriteString("\n")
	}
	
	// Network Status
	b.WriteString("NETWORK STATUS\n")
	for _, iface := range s.NetworkStatus.Interfaces {
		b.WriteString(fmt.Sprintf("  %s: %s\n", iface.Name, iface.Status))
		for _, addr := range iface.Addresses {
			b.WriteString(fmt.Sprintf("    %s\n", addr))
		}
	}
	if len(s.NetworkStatus.DNSServers) > 0 {
		b.WriteString(fmt.Sprintf("  DNS Servers: %s\n", strings.Join(s.NetworkStatus.DNSServers, ", ")))
	}
	b.WriteString("\n")
	
	// Issues Summary
	if len(s.CriticalIssues) > 0 || len(s.Warnings) > 0 {
		b.WriteString("ISSUES DETECTED\n")
		if len(s.CriticalIssues) > 0 {
			b.WriteString(fmt.Sprintf("  Critical Issues: %d\n", len(s.CriticalIssues)))
			for i, issue := range s.CriticalIssues {
				if i < 5 { // Show first 5
					b.WriteString(fmt.Sprintf("    - %s\n", issue))
				}
			}
			if len(s.CriticalIssues) > 5 {
				b.WriteString(fmt.Sprintf("    ... and %d more\n", len(s.CriticalIssues)-5))
			}
		}
		if len(s.Warnings) > 0 {
			b.WriteString(fmt.Sprintf("  Warnings: %d\n", len(s.Warnings)))
			for i, warning := range s.Warnings {
				if i < 5 { // Show first 5
					b.WriteString(fmt.Sprintf("    - %s\n", warning))
				}
			}
			if len(s.Warnings) > 5 {
				b.WriteString(fmt.Sprintf("    ... and %d more\n", len(s.Warnings)-5))
			}
		}
		b.WriteString("\n")
	}
	
	// Recommendations
	if len(s.Recommendations) > 0 {
		b.WriteString("RECOMMENDATIONS\n")
		for i, rec := range s.Recommendations {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, rec))
		}
		b.WriteString("\n")
	}
	
	b.WriteString("=====================================\n")
	b.WriteString("         END OF REPORT              \n")
	b.WriteString("=====================================\n")
	
	return b.String()
}

// Helper functions

func getHealthBar(score int) string {
	filled := score / 10
	bar := "["
	for i := 0; i < 10; i++ {
		if i < filled {
			bar += "#"
		} else {
			bar += "."
		}
	}
	bar += "]"
	return bar
}

func getHealthStatus(score int) string {
	switch {
	case score >= 90:
		return "EXCELLENT"
	case score >= 75:
		return "GOOD"
	case score >= 60:
		return "FAIR"
	case score >= 40:
		return "POOR"
	default:
		return "CRITICAL"
	}
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	}
	return fmt.Sprintf("%d minutes", minutes)
}