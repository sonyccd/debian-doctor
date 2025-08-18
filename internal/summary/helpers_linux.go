// Linux-specific helper functions for system summary generation

package summary

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// getLoadAverage returns the system load average (1, 5, 15 minutes)
func getLoadAverage() ([3]float64, error) {
	var loadAvg [3]float64
	
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return loadAvg, err
	}
	
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return loadAvg, nil
	}
	
	for i := 0; i < 3; i++ {
		if val, err := strconv.ParseFloat(fields[i], 64); err == nil {
			loadAvg[i] = val
		}
	}
	
	return loadAvg, nil
}

// getDNSServers returns the configured DNS servers
func getDNSServers() []string {
	var servers []string
	
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return servers
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				servers = append(servers, fields[1])
			}
		}
	}
	
	return servers
}