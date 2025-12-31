package processes

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/process"
)

func ProcessesInfoCollect(ctx context.Context) (interface{}, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	var processInfos []map[string]interface{}
	for _, proc := range processes {
		ppid, err := proc.Ppid()
		if err != nil {
			return nil, fmt.Errorf("failed to get process info: %w", err)
		}

		cmdline, err := proc.Cmdline()
		if err != nil {
			return nil, fmt.Errorf("failed to get process info: %w", err)
		}

		status, err := proc.Status()
		if err != nil {
			return nil, fmt.Errorf("failed to get process info: %w", err)
		}

		ppidName, err := process.NewProcess(ppid)
		if err != nil {
			return nil, fmt.Errorf("failed to get process info: %w", err)
		}

		cpuUsage, err := proc.CPUPercent()
		if err != nil {
			return nil, fmt.Errorf("failed to get process info: %w", err)
		}

		memoryInfo, err := proc.MemoryInfo()
		if err != nil {
			return nil, fmt.Errorf("failed to get process info: %w", err)
		}

		processInfo := map[string]interface{}{
			"pid":         proc.Pid,
			"ppid":        ppid,
			"status":      status,
			"cmdline":     cmdline,
			"cpuUsage":    cpuUsage,
			"memoryInfo":  memoryInfo,
			"ppidName":    ppidName,
			"ppidStatus":  status,
			"ppidCmdline": cmdline,
		}

		processInfos = append(processInfos, processInfo)
	}

	return map[string]interface{}{
		"processes": processInfos,
	}, nil
}
