package diagnose


import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/debian-doctor/debian-doctor/internal/fixes"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

// DiagnosePerformanceIssues diagnoses performance-related problems
func DiagnosePerformanceIssues() Diagnosis {
	diagnosis := Diagnosis{
		Issue:    "Performance Issues",
		Findings: []string{},
		Fixes:    []*fixes.Fix{},
	}

	// Check CPU usage
	if percent, err := cpu.Percent(0, false); err == nil && len(percent) > 0 {
		cpuUsage := percent[0]
		if cpuUsage > 80 {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("High CPU usage: %.1f%%", cpuUsage))
			
			// Get top CPU processes
			cmd := exec.Command("ps", "aux", "--sort=-pcpu")
			if output, err := cmd.Output(); err == nil {
				lines := strings.Split(string(output), "\n")
				if len(lines) > 1 {
					diagnosis.Findings = append(diagnosis.Findings, "Top CPU consumers:")
					for i := 1; i < 4 && i < len(lines); i++ {
						fields := strings.Fields(lines[i])
						if len(fields) > 10 {
							diagnosis.Findings = append(diagnosis.Findings, 
								fmt.Sprintf("  - %s: %s%% CPU", fields[10], fields[2]))
						}
					}
				}
			}
		} else {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("CPU usage normal: %.1f%%", cpuUsage))
		}
	}

	// Check memory usage
	if vmStat, err := mem.VirtualMemory(); err == nil {
		if vmStat.UsedPercent > 85 {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("High memory usage: %.1f%%", vmStat.UsedPercent))
			
			// Get top memory processes
			cmd := exec.Command("ps", "aux", "--sort=-pmem")
			if output, err := cmd.Output(); err == nil {
				lines := strings.Split(string(output), "\n")
				if len(lines) > 1 {
					diagnosis.Findings = append(diagnosis.Findings, "Top memory consumers:")
					for i := 1; i < 4 && i < len(lines); i++ {
						fields := strings.Fields(lines[i])
						if len(fields) > 10 {
							diagnosis.Findings = append(diagnosis.Findings, 
								fmt.Sprintf("  - %s: %s%% MEM", fields[10], fields[3]))
						}
					}
				}
			}
			
			diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
				ID:           "clear_caches",
				Title:        "Clear System Caches",
				Description:  "Clear system caches",
				Commands:     []string{"sync", "echo 3 > /proc/sys/vm/drop_caches"},
				RequiresRoot: true,
				Reversible:   false,
				RiskLevel:    fixes.RiskLow,
			})
		} else {
			diagnosis.Findings = append(diagnosis.Findings, fmt.Sprintf("Memory usage normal: %.1f%%", vmStat.UsedPercent))
		}
	}

	// Check load average
	if avg, err := load.Avg(); err == nil {
		cpuCount, _ := cpu.Counts(true)
		if avg.Load1 > float64(cpuCount*2) {
			diagnosis.Findings = append(diagnosis.Findings, 
				fmt.Sprintf("High system load: %.2f (cores: %d)", avg.Load1, cpuCount))
			diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
				ID:           "view_processes",
				Title:        "View Running Processes",
				Description:  "View running processes",
				Commands:     []string{"ps aux --sort=-%cpu | head -20"},
				RequiresRoot: false,
				Reversible:   false,
				RiskLevel:    fixes.RiskLow,
			})
		} else {
			diagnosis.Findings = append(diagnosis.Findings, 
				fmt.Sprintf("System load normal: %.2f", avg.Load1))
		}
	}

	// Check for swap usage
	if swapStat, err := mem.SwapMemory(); err == nil {
		if swapStat.UsedPercent > 50 {
			diagnosis.Findings = append(diagnosis.Findings, 
				fmt.Sprintf("High swap usage: %.1f%% - possible memory pressure", swapStat.UsedPercent))
			diagnosis.Fixes = append(diagnosis.Fixes, &fixes.Fix{
				ID:           "clear_swap",
				Title:        "Clear Swap Memory",
				Description:  "Clear swap (requires sufficient RAM)",
				Commands:     []string{"swapoff -a", "swapon -a"},
				RequiresRoot: true,
				Reversible:   false,
				RiskLevel:    fixes.RiskHigh,
			})
		}
	}

	if len(diagnosis.Findings) == 0 {
		diagnosis.Findings = append(diagnosis.Findings, "No performance issues detected")
	}

	return diagnosis
}