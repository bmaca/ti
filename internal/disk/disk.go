package disk

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type DiskInfo struct {
	Total string `json:"total"`
	Used  string `json:"used"`
	Free  string `json:"free"`
}

func GetDiskInfo() DiskInfo {
	info := DiskInfo{
		Total: "",
		Used:  "",
		Free:  "",
	}

	switch runtime.GOOS {
	case "linux":
		info = parseLinuxDisk()
	case "darwin":
		info = parseMacOSDisk()
	case "windows":
		info = parseWindowsDisk()
	}

	return info
}

func parseLinuxDisk() DiskInfo {
	info := DiskInfo{
		Total: "",
		Used:  "",
		Free:  "",
	}

	output, err := exec.Command("df", "-h", "/").Output()
	if err != nil {
		return info
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		parts := strings.Fields(lines[1])

		if len(parts) >= 4 {
			info.Total = parts[1]
			info.Used = parts[2]
			info.Free = parts[3]
		}
	}

	return info
}

func parseMacOSDisk() DiskInfo {
	info := DiskInfo{
		Total: "",
		Used:  "",
		Free:  "",
	}

	output, err := exec.Command("df", "-h", "/").Output()
	if err != nil {
		return info
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		parts := strings.Fields(lines[1])
		if len(parts) >= 4 {
			info.Total = parts[1]
			info.Used = parts[2]
			info.Free = parts[3]
		}
	}

	return info
}

func parseWindowsDisk() DiskInfo {
	info := DiskInfo{
		Total: "",
		Used:  "",
		Free:  "",
	}

	// Get C: drive info
	output, err := exec.Command("wmic", "logicaldisk", "where", "name='C:'", "get", "size,freespace", "/format:list").Output()
	if err != nil {
		return parseWindowsPowerShellDisk()
	}

	lines := strings.Split(string(output), "\n")
	var sizeBytes, freeBytes int64

	for _, line := range lines {
		parts := strings.Split(line, "=")
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Size":
			if sz, err := strconv.ParseInt(value, 10, 64); err == nil {
				sizeBytes = sz
			}
		case "FreeSpace":
			if free, err := strconv.ParseInt(value, 10, 64); err == nil {
				freeBytes = free
			}
		}
	}

	if sizeBytes > 0 {
		info.Total = formatBytes(sizeBytes)
		info.Free = formatBytes(freeBytes)
		usedBytes := sizeBytes - freeBytes
		info.Used = formatBytes(usedBytes)
	}

	return info
}

func parseWindowsPowerShellDisk() DiskInfo {
	info := DiskInfo{
		Total: "",
		Used:  "",
		Free:  "",
	}
	var sizeBytes int64
	cmd := `(Get-Volume -DriveLetter C).Size; (Get-Volume -DriveLetter C).SizeRemaining`
	output, err := exec.Command("powershell", "-NoProfile", "-Command", cmd).Output()
	if err != nil {
		return info
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) >= 2 {
		if sizeBytes, err := strconv.ParseInt(strings.TrimSpace(lines[0]), 10, 64); err == nil {
			info.Total = formatBytes(sizeBytes)
		}
		if freeBytes, err := strconv.ParseInt(strings.TrimSpace(lines[1]), 10, 64); err == nil {
			info.Free = formatBytes(freeBytes)
			usedBytes := sizeBytes - freeBytes
			info.Used = formatBytes(usedBytes)
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
