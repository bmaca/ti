package memory

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type MemInfo struct {
	Total     string `json:"total"`
	Available string `json:"available"`
	Used      string `json:"used"`
}

func GetMemoryInfo() MemInfo {
	info := MemInfo{
		Total:     "Unknown",
		Available: "Unknown",
		Used:      "Unknown",
	}

	switch runtime.GOOS {
	case "linux":
		info = parseLinuxMemory()
	case "darwin":
		info = parseMacOSMemory()
	case "windows":
		info = parseWindowsMemory()
	}

	return info
}

func parseLinuxMemory() MemInfo {
	info := MemInfo{
		Total:     "Unknown",
		Available: "Unknown",
		Used:      "Unknown",
	}

	output, err := exec.Command("free", "-h").Output()
	if err != nil {
		return info
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		memLine := lines[1]
		parts := strings.Fields(memLine)

		if len(parts) >= 7 {
			info.Total = parts[1]
			info.Used = parts[2]
			info.Available = parts[6]
		}
	}

	return info
}

func parseMacOSMemory() MemInfo {
	info := MemInfo{
		Total:     "Unknown",
		Available: "Unknown",
		Used:      "Unknown",
	}

	if output, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
		if totalBytes, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64); err == nil {
			info.Total = formatBytes(totalBytes)
		}
	}

	if output, err := exec.Command("vm_stat").Output(); err == nil {
		info.Available, info.Used = parseVMStat(string(output), info.Total)
	}

	return info
}

func parseVMStat(output, total string) (available, used string) {
	lines := strings.Split(output, "\n")
	var pagesFree int64 = 0
	var pagesInactive int64 = 0
	var pageSize int64 = 4096 // Default page size

	if psizeOut, err := exec.Command("sysctl", "-n", "hw.pagesize").Output(); err == nil {
		if psize, err := strconv.ParseInt(strings.TrimSpace(string(psizeOut)), 10, 64); err == nil {
			pageSize = psize
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			if strings.HasPrefix(line, "Pages free:") {
				if pf, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
					pagesFree = pf
				}
			}
			if strings.HasPrefix(line, "Pages inactive:") {
				if pi, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
					pagesInactive = pi
				}
			}
		}
	}

	availableBytes := (pagesFree + pagesInactive) * pageSize
	available = formatBytes(availableBytes)
	return available, "N/A (use Activity Monitor)"
}

func parseWindowsMemory() MemInfo {
	info := MemInfo{
		Total:     "Unknown",
		Available: "Unknown",
		Used:      "Unknown",
	}

	output, err := exec.Command("wmic", "OS", "get", "TotalVisibleMemorySize,FreePhysicalMemory", "/format:list").Output()
	if err != nil {
		return parseWindowsPowerShellMemory()
	}

	lines := strings.Split(string(output), "\n")
	var totalKB, freeKB int64

	for _, line := range lines {
		parts := strings.Split(line, "=")
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "TotalVisibleMemorySize":
			if kb, err := strconv.ParseInt(value, 10, 64); err == nil {
				totalKB = kb
			}
		case "FreePhysicalMemory":
			if kb, err := strconv.ParseInt(value, 10, 64); err == nil {
				freeKB = kb
			}
		}
	}

	if totalKB > 0 {
		info.Total = formatBytes(totalKB * 1024)
		info.Available = formatBytes(freeKB * 1024)
		usedKB := totalKB - freeKB
		info.Used = formatBytes(usedKB * 1024)
	}

	return info
}

func parseWindowsPowerShellMemory() MemInfo {
	info := MemInfo{
		Total:     "Unknown",
		Available: "Unknown",
		Used:      "Unknown",
	}
	var totalKB int64
	cmd := `(Get-CimInstance -ClassName Win32_OperatingSystem).TotalVisibleMemorySize; (Get-CimInstance -ClassName Win32_OperatingSystem).FreePhysicalMemory`
	output, err := exec.Command("powershell", "-NoProfile", "-Command", cmd).Output()
	if err != nil {
		return info
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) >= 2 {
		if totalKB, err := strconv.ParseInt(strings.TrimSpace(lines[0]), 10, 64); err == nil {
			info.Total = formatBytes(totalKB * 1024)
		}
		if freeKB, err := strconv.ParseInt(strings.TrimSpace(lines[1]), 10, 64); err == nil {
			info.Available = formatBytes(freeKB * 1024)
			usedKB := strconv.FormatInt(totalKB-freeKB, 10)
			if ub, err := strconv.ParseInt(usedKB, 10, 64); err == nil {
				info.Used = formatBytes(ub * 1024)
			}
		}
	}

	return info
}

func formatBytes(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	divisor := int64(1024)
	value := float64(bytes)

	for _, unit := range units {
		if value < float64(divisor) {
			if unit == "B" {
				return strconv.FormatInt(int64(value), 10) + " " + unit
			}
			return strconv.FormatFloat(value, 'f', 2, 64) + " " + unit
		}
		value /= float64(divisor)
	}

	return strconv.FormatFloat(value, 'f', 2, 64) + " PB"
}
