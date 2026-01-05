package cpu

import (
	"context"
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
)

func CPUInfo(ctx context.Context) (interface{}, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	// Get CPU usage (average over 1 second)
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	var usage float64
	if len(percentages) > 0 {
		usage = percentages[0]
	}

	result := map[string]interface{}{
		"usage_percent": usage,
		"cores":         runtime.NumCPU(),
	}

	if len(cpuInfo) > 0 {
		result["model"] = cpuInfo[0].ModelName
		result["mhz"] = cpuInfo[0].Mhz
		result["vendor"] = cpuInfo[0].VendorID
	}

	return result, nil
}
