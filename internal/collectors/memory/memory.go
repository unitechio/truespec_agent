package memory

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/mem"
)

func MemoryInfo(ctx context.Context) (interface{}, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	return map[string]interface{}{
		"total_mb":     vmStat.Total / 1024 / 1024,
		"available_mb": vmStat.Available / 1024 / 1024,
		"used_mb":      vmStat.Used / 1024 / 1024,
		"used_percent": vmStat.UsedPercent,
		"free_mb":      vmStat.Free / 1024 / 1024,
		"cached_mb":    vmStat.Cached / 1024 / 1024,
		"buffers_mb":   vmStat.Buffers / 1024 / 1024,
	}, nil
}
