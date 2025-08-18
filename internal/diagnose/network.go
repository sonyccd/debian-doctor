package diagnose


import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/debian-doctor/debian-doctor/internal/fixes"
)

// DiagnoseNetworkIssues diagnoses network-related problems
func DiagnoseNetworkIssues() Diagnosis {
	diagnosis := Diagnosis{
		Issue:    "Network Issues",
		Findings: []string{},
		Fixes:    []*fixes.Fix{},
	}

	// Check networking service
	if output, err := exec.Command("systemctl", "is-active", "networking").Output(); err != nil {
		diagnosis.Findings = append(diagnosis.Findings, "Networking service is not running")
		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:           "restart_networking",
			Title:        "Restart Networking Service",
			Description:  "Restart networking service",
			Commands:     []string{"systemctl restart networking"},
			RequiresRoot: true,
			Reversible:   false,
			RiskLevel:    fixes.RiskMedium,
		})
	} else if strings.TrimSpace(string(output)) == "active" {
		diagnosis.Findings = append(diagnosis.Findings, "Networking service is active")
	}

	// Check interfaces
	interfaces, err := net.Interfaces()
	if err == nil {
		downInterfaces := []string{}
		for _, iface := range interfaces {
			if iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			
			if iface.Flags&net.FlagUp == 0 {
				downInterfaces = append(downInterfaces, iface.Name)
			}
		}
		
		if len(downInterfaces) > 0 {
			diagnosis.Findings = append(diagnosis.Findings, 
				fmt.Sprintf("Interfaces down: %s", strings.Join(downInterfaces, ", ")))
			for _, iface := range downInterfaces {
				diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
					ID:           fmt.Sprintf("bring_up_%s", iface),
					Title:        fmt.Sprintf("Bring Up Interface %s", iface),
					Description:  fmt.Sprintf("Bring up interface %s", iface),
					Commands:     []string{fmt.Sprintf("ip link set %s up", iface)},
					RequiresRoot: true,
					Reversible:   true,
					ReverseCommands: []string{fmt.Sprintf("ip link set %s down", iface)},
					RiskLevel:    fixes.RiskMedium,
				})
			}
		} else {
			diagnosis.Findings = append(diagnosis.Findings, "All network interfaces are up")
		}
	}

	// Check DNS resolution
	if _, err := net.LookupHost("debian.org"); err != nil {
		diagnosis.Findings = append(diagnosis.Findings, "DNS resolution failed")
		diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
			ID:           "reset_dns",
			Title:        "Reset DNS Configuration",
			Description:  "Reset DNS configuration",
			Commands:     []string{"echo 'nameserver 8.8.8.8' > /etc/resolv.conf"},
			RequiresRoot: true,
			Reversible:   false,
			RiskLevel:    fixes.RiskHigh,
		})
	} else {
		diagnosis.Findings = append(diagnosis.Findings, "DNS resolution working")
	}

	// Check default route
	if output, err := exec.Command("ip", "route", "show", "default").Output(); err == nil {
		if len(output) == 0 {
			diagnosis.Findings = append(diagnosis.Findings, "No default route configured")
			diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
				ID:           "add_default_route",
				Title:        "Add Default Route",
				Description:  "Add default route (replace IP with your gateway)",
				Commands:     []string{"ip route add default via 192.168.1.1"},
				RequiresRoot: true,
				Reversible:   true,
				ReverseCommands: []string{"ip route del default via 192.168.1.1"},
				RiskLevel:    fixes.RiskHigh,
			})
		} else {
			diagnosis.Findings = append(diagnosis.Findings, "Default route configured")
		}
	}

	if len(diagnosis.Findings) == 0 {
		diagnosis.Findings = append(diagnosis.Findings, "No network issues detected")
	}

	return diagnosis
}