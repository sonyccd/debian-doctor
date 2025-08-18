package checks

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

// MemoryCheck checks memory usage
type MemoryCheck struct{}

func (m MemoryCheck) Name() string {
	return "Memory Usage"
}

func (m MemoryCheck) RequiresRoot() bool {
	return false
}

func (m MemoryCheck) Run() CheckResult {
	result := CheckResult{
		Name:      m.Name(),
		Severity:  SeverityInfo,
		Timestamp: time.Now(),
		Details:   []string{},
	}

	// Get virtual memory stats
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		result.Severity = SeverityError
		result.Message = "Failed to check memory usage"
		return result
	}

	// Format memory information
	result.Details = append(result.Details, fmt.Sprintf("Total: %d MB", vmStat.Total/(1024*1024)))
	result.Details = append(result.Details, fmt.Sprintf("Available: %d MB", vmStat.Available/(1024*1024)))
	result.Details = append(result.Details, fmt.Sprintf("Used: %d MB (%.1f%%)", vmStat.Used/(1024*1024), vmStat.UsedPercent))

	// Check memory usage severity
	switch {
	case vmStat.UsedPercent > 90:
		result.Severity = SeverityError
		result.Message = fmt.Sprintf("Memory usage critical: %.1f%%", vmStat.UsedPercent)
	case vmStat.UsedPercent > 80:
		result.Severity = SeverityWarning
		result.Message = fmt.Sprintf("Memory usage high: %.1f%%", vmStat.UsedPercent)
	default:
		result.Severity = SeverityInfo
		result.Message = fmt.Sprintf("Memory usage OK: %.1f%%", vmStat.UsedPercent)
	}

	// Check swap usage
	swapStat, err := mem.SwapMemory()
	if err == nil {
		result.Details = append(result.Details, fmt.Sprintf("Swap Total: %d MB", swapStat.Total/(1024*1024)))
		result.Details = append(result.Details, fmt.Sprintf("Swap Used: %d MB (%.1f%%)", swapStat.Used/(1024*1024), swapStat.UsedPercent))
		
		if swapStat.Total == 0 {
			result.Details = append(result.Details, "Warning: No swap space configured")
		} else if swapStat.UsedPercent > 50 {
			result.Severity = SeverityWarning
			result.Message += " (High swap usage indicates memory pressure)"
		}
	}

	return result
}