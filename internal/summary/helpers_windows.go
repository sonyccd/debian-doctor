//go:build windows

package summary

// getLoadAverage returns empty load average on Windows
func getLoadAverage() ([3]float64, error) {
	return [3]float64{}, nil
}

// getDNSServers returns empty DNS servers on Windows (not implemented)
func getDNSServers() []string {
	return []string{}
}