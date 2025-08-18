package checks

import (
	"fmt"
	"syscall"
	"time"
)

// DiskSpaceCheck checks disk space usage
type DiskSpaceCheck struct{}

func (d DiskSpaceCheck) Name() string {
	return "Disk Space"
}

func (d DiskSpaceCheck) RequiresRoot() bool {
	return false
}

func (d DiskSpaceCheck) Run() CheckResult {
	result := CheckResult{
		Name:      d.Name(),
		Severity:  SeverityInfo,
		Timestamp: time.Now(),
		Details:   []string{},
	}

	// Check main filesystem
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		result.Severity = SeverityError
		result.Message = "Failed to check disk space"
		return result
	}

	// Calculate usage percentage
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	usagePercent := int((used * 100) / total)

	result.Details = append(result.Details, fmt.Sprintf("Total: %d GB", total/(1024*1024*1024)))
	result.Details = append(result.Details, fmt.Sprintf("Used: %d GB (%d%%)", used/(1024*1024*1024), usagePercent))
	result.Details = append(result.Details, fmt.Sprintf("Free: %d GB", free/(1024*1024*1024)))

	// Set severity based on usage
	switch {
	case usagePercent > 95:
		result.Severity = SeverityCritical
		result.Message = fmt.Sprintf("Disk usage critical: %d%%", usagePercent)
	case usagePercent > 85:
		result.Severity = SeverityWarning
		result.Message = fmt.Sprintf("Disk usage high: %d%%", usagePercent)
	default:
		result.Severity = SeverityInfo
		result.Message = fmt.Sprintf("Disk usage OK: %d%%", usagePercent)
	}

	// Check inode usage
	inodeTotal := stat.Files
	inodeFree := stat.Ffree
	inodeUsed := inodeTotal - inodeFree
	inodeUsagePercent := int((inodeUsed * 100) / inodeTotal)
	
	result.Details = append(result.Details, fmt.Sprintf("Inode usage: %d%%", inodeUsagePercent))
	
	if inodeUsagePercent > 90 {
		result.Severity = SeverityWarning
		result.Message += fmt.Sprintf(" (High inode usage: %d%%)", inodeUsagePercent)
	}

	return result
}