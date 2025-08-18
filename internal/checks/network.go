package checks

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// NetworkCheck checks network configuration
type NetworkCheck struct{}

func (n NetworkCheck) Name() string {
	return "Network Configuration"
}

func (n NetworkCheck) RequiresRoot() bool {
	return false
}

func (n NetworkCheck) Run() CheckResult {
	result := CheckResult{
		Name:      n.Name(),
		Severity:  SeverityInfo,
		Timestamp: time.Now(),
		Details:   []string{},
	}

	// Check network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		result.Severity = SeverityError
		result.Message = "Failed to check network interfaces"
		return result
	}

	hasActiveInterface := false
	for _, iface := range interfaces {
		// Skip loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Check if interface is up
		if iface.Flags&net.FlagUp != 0 {
			hasActiveInterface = true
			result.Details = append(result.Details, fmt.Sprintf("Interface %s is UP", iface.Name))
			
			// Get addresses for this interface
			addrs, err := iface.Addrs()
			if err == nil && len(addrs) > 0 {
				for _, addr := range addrs {
					result.Details = append(result.Details, fmt.Sprintf("  IP: %s", addr.String()))
				}
			} else {
				result.Details = append(result.Details, fmt.Sprintf("  No IP address assigned to %s", iface.Name))
			}
		} else {
			result.Details = append(result.Details, fmt.Sprintf("Interface %s is DOWN", iface.Name))
		}
	}

	// Check DNS configuration
	if resolvConf, err := os.ReadFile("/etc/resolv.conf"); err == nil {
		lines := strings.Split(string(resolvConf), "\n")
		dnsServers := []string{}
		for _, line := range lines {
			if strings.HasPrefix(line, "nameserver") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					dnsServers = append(dnsServers, parts[1])
				}
			}
		}
		
		if len(dnsServers) > 0 {
			result.Details = append(result.Details, fmt.Sprintf("DNS servers: %s", strings.Join(dnsServers, ", ")))
		} else {
			result.Details = append(result.Details, "No DNS servers configured")
			result.Severity = SeverityWarning
		}
	}

	// Set overall result
	if !hasActiveInterface {
		result.Severity = SeverityError
		result.Message = "No active network interfaces found"
	} else {
		result.Message = "Network configuration OK"
	}

	return result
}