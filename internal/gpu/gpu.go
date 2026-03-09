package gpu

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type GPUInfo struct {
	Name   string `json:"name"`
	Memory string `json:"memory"`
}

func GetGPUInfo() GPUInfo {
	info := GPUInfo{
		Name:   "Not Detected",
		Memory: "N/A",
	}

	switch runtime.GOOS {
	case "linux":
		info = getLinuxGPU()
	case "darwin":
		info = getMacOSGPU()
	case "windows":
		info = getWindowsGPU()
	}

	return info
}

func getLinuxGPU() GPUInfo {
	info := GPUInfo{
		Name:   "Not Detected",
		Memory: "N/A",
	}

	if output, err := exec.Command("nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader").Output(); err == nil {
		parts := strings.Split(strings.TrimSpace(string(output)), ",")
		if len(parts) >= 2 {
			info.Name = strings.TrimSpace(parts[0])
			info.Memory = strings.TrimSpace(parts[1])
			return info
		}
	}

	if _, err := exec.Command("rocm-smi", "--showproductname").Output(); err == nil {
		info.Name = "AMD Radeon (ROCm)"
		info.Memory = "Check rocm-smi for details"
		return info
	}

	if output, err := exec.Command("clinfo").Output(); err == nil {
		if strings.Contains(string(output), "Intel") {
			info.Name = "Intel Arc"
			return info
		}
	}

	if output, err := exec.Command("lspci").Output(); err == nil {
		info = parseLSPCI(string(output))
	}

	return info
}

func getMacOSGPU() GPUInfo {
	info := GPUInfo{
		Name:   "Not Detected",
		Memory: "N/A",
	}

	output, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
	if err != nil {
		return info
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "Chipset Model:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				info.Name = strings.TrimSpace(parts[1])
			}
		}

		if strings.Contains(line, "VRAM (Dynamic):") || strings.Contains(line, "VRAM (Total):") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				info.Memory = strings.TrimSpace(parts[1])
			}
		}

		if strings.Contains(line, "Apple") && (strings.Contains(line, "M1") || strings.Contains(line, "M2") || strings.Contains(line, "M3")) {
			info.Name = strings.TrimSpace(line)
		}
	}

	return info
}

func getWindowsGPU() GPUInfo {
	info := GPUInfo{
		Name:   "Not Detected",
		Memory: "N/A",
	}

	output, err := exec.Command("wmic", "path", "win32_videocontroller", "get", "name,adapterram", "/format:list").Output()
	if err != nil {
		return getWindowsPowerShellGPU()
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "=")
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			if value != "" && info.Name == "Not Detected" {
				info.Name = value
			}
		case "AdapterRAM":
			if value != "" && value != "0" {
				info.Memory = formatGPUMemory(value)
			}
		}
	}

	return info
}

func getWindowsPowerShellGPU() GPUInfo {
	info := GPUInfo{
		Name:   "Not Detected",
		Memory: "N/A",
	}

	cmd := `Get-CimInstance -ClassName Win32_VideoController | Select-Object -Property Name, AdapterRAM`
	output, err := exec.Command("powershell", "-NoProfile", "-Command", cmd).Output()
	if err != nil {
		return info
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "Name") && !strings.Contains(line, "--") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				info.Name = line
			}
		}
	}

	return info
}

func parseLSPCI(output string) GPUInfo {
	info := GPUInfo{
		Name:   "Not Detected",
		Memory: "N/A",
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "VGA") || strings.Contains(line, "3D") {
			parts := strings.Split(line, ": ")
			if len(parts) >= 2 {
				info.Name = strings.TrimSpace(parts[1])
			}
			break
		}
	}

	return info
}

func formatGPUMemory(bytesStr string) string {
	var bytes int64
	if _, err := fmt.Sscanf(bytesStr, "%d", &bytes); err != nil {
		if b, err := strconv.ParseInt(bytesStr, 10, 64); err == nil {
			bytes = b
		} else {
			return bytesStr
		}
	}

	if bytes == 0 {
		return "N/A"
	}

	gb := float64(bytes) / (1024 * 1024 * 1024)
	if gb >= 1 {
		return strconv.FormatFloat(gb, 'f', 1, 64) + " GB"
	}

	mb := float64(bytes) / (1024 * 1024)
	return strconv.FormatFloat(mb, 'f', 1, 64) + " MB"
}
