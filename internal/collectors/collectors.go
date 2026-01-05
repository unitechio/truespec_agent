package collectors

import (
	"context"

	"github.com/unitechio/agent/internal/collectors/cpu"
	"github.com/unitechio/agent/internal/collectors/disk"
	"github.com/unitechio/agent/internal/collectors/memory"
	"github.com/unitechio/agent/internal/collectors/network"
	"github.com/unitechio/agent/internal/collectors/processes"
	"github.com/unitechio/agent/internal/collectors/system"
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
	return system.OSInfo(ctx)
}

type CPUCollector struct{}

func (c *CPUCollector) Name() string {
	return "cpu"
}

func (c *CPUCollector) Collect(ctx context.Context) (interface{}, error) {
	return cpu.CPUInfo(ctx)
}

type MemoryCollector struct{}

func (c *MemoryCollector) Name() string {
	return "memory"
}

func (c *MemoryCollector) Collect(ctx context.Context) (interface{}, error) {
	return memory.MemoryInfo(ctx)
}

type DiskCollector struct{}

func (c *DiskCollector) Name() string {
	return "disk"
}

func (c *DiskCollector) Collect(ctx context.Context) (interface{}, error) {
	return disk.DiskInfo(ctx)
}

type NetworkCollector struct {
	CollectMAC bool // Policy-controlled: whether to collect MAC addresses
}

func (c *NetworkCollector) Name() string {
	return "network"
}

func (c *NetworkCollector) Collect(ctx context.Context) (interface{}, error) {
	return network.NetworkInfo(ctx, c.CollectMAC)
}

type ProcessesCollector struct{}

func (p *ProcessesCollector) Name() string {
	return "processes"
}

func (f *ProcessesCollector) Collect(ctx context.Context) (interface{}, error) {
	return processes.ProcessesInfoCollect(ctx)
}

type FileCollector struct {
}

func (f *FileCollector) Name() string {
	return "file"
}

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
