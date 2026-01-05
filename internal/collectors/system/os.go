package system

import (
	"context"
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v3/host"
)

func OSInfo(ctx context.Context) (interface{}, error) {
	info, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	return map[string]interface{}{
		"hostname":         info.Hostname,
		"os":               info.OS,
		"platform":         info.Platform,
		"platform_family":  info.PlatformFamily,
		"platform_version": info.PlatformVersion,
		"kernel_version":   info.KernelVersion,
		"kernel_arch":      info.KernelArch,
		"uptime":           info.Uptime,
		"boot_time":        info.BootTime,
		"go_version":       runtime.Version(),
		"num_cpu":          runtime.NumCPU(),
	}, nil
}
