package disk

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/disk"
)

func DiskInfo(ctx context.Context) (interface{}, error) {
	partitions, err := disk.Partitions(false) // false = exclude pseudo filesystems
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	var disks []map[string]interface{}
	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		disks = append(disks, map[string]interface{}{
			"device":       partition.Device,
			"mountpoint":   partition.Mountpoint,
			"fstype":       partition.Fstype,
			"total_gb":     usage.Total / 1024 / 1024 / 1024,
			"used_gb":      usage.Used / 1024 / 1024 / 1024,
			"free_gb":      usage.Free / 1024 / 1024 / 1024,
			"used_percent": usage.UsedPercent,
		})
	}

	return map[string]interface{}{
		"partitions": disks,
	}, nil
}
