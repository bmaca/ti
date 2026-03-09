package cpu

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type CPUInfo struct {
	Brand string `json:"brand"`
	Model string `json:"model"`
	Cores int    `json:"cores"`
	Speed string `json:"speed"`
}

func GetCPUInfo() CPUInfo {
	info := CPUInfo{
		Brand: "Unknown",
		Model: "Unknown",
		Cores: 0,
		Speed: "Unknown",
	}

	info.Cores = runtime.NumCPU()

	switch runtime.GOOS {
	case "linux":
		info = parseLSCPU(getLinuxCPUInfo())
	case "darwin":
		info = parseMacOSCPUInfo(info)
	case "windows":
		info = parseWindowsCPUInfo(info)
	}

	if info.Cores == 0 {
		info.Cores = runtime.NumCPU()
	}

	return info
}

func getLinuxCPUInfo() string {
	output, err := exec.Command("lscpu").Output()
	if err != nil {
		return ""
	}
	return string(output)
}

func parseLSCPU(output string) CPUInfo {
	info := CPUInfo{
		Brand: "Unknown",
		Model: "Unknown",
		Cores: 0,
		Speed: "Unknown",
	}

	if output == "" {
		return info
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Model name":
			if strings.Contains(value, "Intel") {
				info.Brand = "Intel"
				info.Model = strings.TrimPrefix(value, "Intel(R) Core(TM) ")
				info.Model = strings.TrimPrefix(info.Model, "Intel(R) ")
			} else if strings.Contains(value, "AMD") {
				info.Brand = "AMD"
				info.Model = strings.TrimPrefix(value, "AMD Ryzen ")
				info.Model = strings.TrimPrefix(info.Model, "AMD ")
			} else {
				info.Model = value
			}
		case "CPU(s)":
			if cores, err := strconv.Atoi(value); err == nil {
				info.Cores = cores
			}
		case "CPU max MHz":
			info.Speed = value + " MHz"
		}
	}

	return info
}

func parseMacOSCPUInfo(info CPUInfo) CPUInfo {
	if output, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
		modelStr := strings.TrimSpace(string(output))

		if strings.Contains(modelStr, "Intel") {
			info.Brand = "Intel"
			info.Model = strings.TrimPrefix(modelStr, "Intel(R) Core(TM) ")
			info.Model = strings.TrimPrefix(info.Model, "Intel(R) ")
		} else if strings.Contains(modelStr, "Apple") {
			info.Brand = "Apple"
			info.Model = strings.TrimSpace(modelStr)
		} else {
			info.Model = modelStr
		}
	}

	if output, err := exec.Command("sysctl", "-n", "hw.physicalcpu").Output(); err == nil {
		if cores, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
			info.Cores = cores
		}
	}

	if output, err := exec.Command("sysctl", "-n", "hw.cpufrequency").Output(); err == nil {
		freqStr := strings.TrimSpace(string(output))
		if freq, err := strconv.Atoi(freqStr); err == nil {
			info.Speed = strconv.Itoa(freq/1000000) + " MHz"
		}
	}

	return info
}

func parseWindowsCPUInfo(info CPUInfo) CPUInfo {
	// Use WMI to get CPU info
	// wmic cpu get name, cores
	output, err := exec.Command("wmic", "cpu", "get", "name,numberOfCores,maxClockSpeed", "/format:list").Output()
	if err != nil {
		// Fallback: try powershell
		return parseWindowsPowerShellCPUInfo(info)
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
			if strings.Contains(value, "Intel") {
				info.Brand = "Intel"
			} else if strings.Contains(value, "AMD") {
				info.Brand = "AMD"
			}
			info.Model = value
		case "MaxClockSpeed":
			info.Speed = value + " MHz"
		case "NumberOfCores":
			if cores, err := strconv.Atoi(value); err == nil {
				info.Cores = cores
			}
		}
	}

	return info
}

func parseWindowsPowerShellCPUInfo(info CPUInfo) CPUInfo {
	cmd := `Get-CimInstance -ClassName Win32_Processor | Select-Object -Property Name,NumberOfCores,MaxClockSpeed`
	output, err := exec.Command("powershell", "-NoProfile", "-Command", cmd).Output()
	if err != nil {
		return info
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) > 0 && !strings.Contains(line, "Name") && !strings.Contains(line, "--") {
			if strings.Contains(line, "Intel") {
				info.Brand = "Intel"
				info.Model = line
			} else if strings.Contains(line, "AMD") {
				info.Brand = "AMD"
				info.Model = line
			}
		}
	}

	return info
}
