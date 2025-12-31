package collectors

import (
	"context"
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/unitechio/agent/internal/collectors/processes"
)

type Collector interface {
	Name() string
	Collect(ctx context.Context) (interface{}, error)
}

type SystemCollector struct{}

func (c *SystemCollector) Name() string {
	return "system"
}

func (c *SystemCollector) Collect(ctx context.Context) (interface{}, error) {
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

type CPUCollector struct{}

func (c *CPUCollector) Name() string {
	return "cpu"
}

func (c *CPUCollector) Collect(ctx context.Context) (interface{}, error) {
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

type MemoryCollector struct{}

func (c *MemoryCollector) Name() string {
	return "memory"
}

func (c *MemoryCollector) Collect(ctx context.Context) (interface{}, error) {
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

type DiskCollector struct{}

func (c *DiskCollector) Name() string {
	return "disk"
}

func (c *DiskCollector) Collect(ctx context.Context) (interface{}, error) {
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

type NetworkCollector struct {
	CollectMAC bool // Policy-controlled: whether to collect MAC addresses
}

func (c *NetworkCollector) Name() string {
	return "network"
}

func (c *NetworkCollector) Collect(ctx context.Context) (interface{}, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var networks []map[string]interface{}
	for _, iface := range interfaces {
		netInfo := map[string]interface{}{
			"name":  iface.Name,
			"mtu":   iface.MTU,
			"flags": iface.Flags,
		}

		// Add IP addresses
		var addrs []string
		for _, addr := range iface.Addrs {
			addrs = append(addrs, addr.Addr)
		}
		netInfo["addresses"] = addrs

		// Only include MAC address if policy allows
		if c.CollectMAC {
			netInfo["mac"] = iface.HardwareAddr
		}

		networks = append(networks, netInfo)
	}

	return map[string]interface{}{
		"interfaces": networks,
	}, nil
}

type ProcessesCollector struct{}

func (p *ProcessesCollector) Name() string {
	return "processes"
}

func (f *ProcessesCollector) Collect(ctx context.Context) (interface{}, error) {
	return processes.ProcessesInfoCollect(ctx)
}

// type FileCollector struct {
// }

// func (f *FileCollector) Name() string {
// 	return "file"
// }

// func (f *FileCollector) Collect(ctx context.Context, os string) (interface{}, error) {

// }

func NewDefaultCollectors() []Collector {
	return []Collector{
		&SystemCollector{},
		&CPUCollector{},
		&MemoryCollector{},
		&DiskCollector{},
		// &FileCollector{},
		&ProcessesCollector{},
		&NetworkCollector{CollectMAC: false}, // MAC collection disabled by default
	}
}
