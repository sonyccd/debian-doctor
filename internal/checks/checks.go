package checks

import "os"

// GetAllChecks returns all available system checks
func GetAllChecks() []Check {
	isRoot := os.Geteuid() == 0
	
	checks := []Check{
		SystemInfoCheck{},
		DiskSpaceCheck{},
		MemoryCheck{},
		NetworkCheck{},
		LogsCheck{},
		PackagesCheck{},
		FilesystemCheck{},
	}
	
	// Add root-only checks if running as root
	if isRoot {
		checks = append(checks, ServicesCheck{})
	}
	
	return checks
}