package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/bmaca/ti/internal/gpu"

	"github.com/bmaca/ti/internal/memory"

	"github.com/bmaca/ti/internal/disk"

	"github.com/bmaca/ti/internal/cpu"
)

type SystemInfo struct {
	CPU    cpu.CPUInfo    `json:"cpu"`
	Memory memory.MemInfo `json:"memory"`
	Disk   disk.DiskInfo  `json:"disk"`
	GPU    gpu.GPUInfo    `json:"gpu"`
	OS     OSInfo         `json:"os"`
}

type OSInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

func main() {
	outputFormat := flag.String("format", "table", "Output format: table, json, or text")
	flag.Parse()

	sysInfo := SystemInfo{
		CPU:    cpu.GetCPUInfo(),
		Memory: memory.GetMemoryInfo(),
		Disk:   disk.GetDiskInfo(),
		GPU:    gpu.GetGPUInfo(),
		OS: OSInfo{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	}

	switch *outputFormat {
	case "json":
		outputJSON(sysInfo)
	case "text":
		outputText(sysInfo)
	case "table":
		outputTable(sysInfo)
	default:
		fmt.Fprintf(os.Stderr, "Invalid format: %s. Use: table, json, or text\n", *outputFormat)
		os.Exit(1)
	}
}

func outputTable(info SystemInfo) {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║            SYSTEM SPECIFICATIONS                           ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")

	// CPU
	fmt.Printf("║ %-28s │ %-28s ║\n", "CPU", info.CPU.Brand+" "+info.CPU.Model)
	fmt.Printf("║ %-28s │ %-28s ║\n", "CPU Cores", fmt.Sprintf("%d cores", info.CPU.Cores))
	fmt.Printf("║ %-28s │ %-28s ║\n", "CPU Speed", info.CPU.Speed)

	// Architecture
	fmt.Printf("║ %-28s │ %-28s ║\n", "Architecture", info.OS.Arch)

	// Memory
	fmt.Printf("║ %-28s │ %-28s ║\n", "Memory", info.Memory.Total)
	fmt.Printf("║ %-28s │ %-28s ║\n", "Memory Available", info.Memory.Available)

	// Disk
	if info.Disk.Total != "" {
		fmt.Printf("║ %-28s │ %-28s ║\n", "Storage", info.Disk.Total)
		fmt.Printf("║ %-28s │ %-28s ║\n", "Storage Used", info.Disk.Used)
	}

	// GPU
	if info.GPU.Name != "Not Detected" {
		fmt.Printf("║ %-28s │ %-28s ║\n", "GPU", info.GPU.Name)
		fmt.Printf("║ %-28s │ %-28s ║\n", "GPU Memory", info.GPU.Memory)
	}

	// OS
	fmt.Printf("║ %-28s │ %-28s ║\n", "OS", info.OS.OS)

	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}

func outputText(info SystemInfo) {
	fmt.Printf("CPU: %s %s (%d cores)\n", info.CPU.Brand, info.CPU.Model, info.CPU.Cores)
	fmt.Printf("Architecture: %s\n", info.OS.Arch)
	fmt.Printf("Memory: %s (Available: %s)\n", info.Memory.Total, info.Memory.Available)
	if info.Disk.Total != "" {
		fmt.Printf("Storage: %s (Used: %s)\n", info.Disk.Total, info.Disk.Used)
	}
	if info.GPU.Name != "Not Detected" {
		fmt.Printf("GPU: %s (%s)\n", info.GPU.Name, info.GPU.Memory)
	}
	fmt.Printf("OS: %s\n", info.OS.OS)
}

func outputJSON(info SystemInfo) {
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}
