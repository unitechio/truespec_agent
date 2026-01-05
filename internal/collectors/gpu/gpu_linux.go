//go:build linux

package gpu

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/jaypipes/ghw"
)

func getGPUs() ([]GPUInfo, error) {
	info, err := ghw.GPU()
	if err != nil {
		return nil, err
	}

	gpus := make([]GPUInfo, 0)

	for _, card := range info.GraphicsCards {
		gpu := GPUInfo{
			Name:   card.DeviceInfo.Product.Name,
			Vendor: card.DeviceInfo.Vendor.Name,
		}

		if strings.Contains(strings.ToLower(gpu.Vendor), "nvidia") {
			enrichNvidia(&gpu)
		}

		gpus = append(gpus, gpu)
	}

	return gpus, nil
}

func enrichNvidiaAll(gpus []GPUInfo) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"nvidia-smi",
		"--query-gpu=index,name,pcie.bus_id,memory.total,memory.used,driver_version",
		"--format=csv,noheader,nounits",
	)

	out, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		f := strings.Split(line, ",")
		if len(f) < 6 {
			continue
		}

		idx := int(parseUint(f[0]))
		if idx >= len(gpus) {
			continue
		}

		gpus[idx].Name = strings.TrimSpace(f[1])
		gpus[idx].BusID = strings.TrimSpace(f[2])
		gpus[idx].VRAMTotalMB = parseUint(f[3])
		gpus[idx].VRAMUsedMB = parseUint(f[4])
		gpus[idx].Driver = strings.TrimSpace(f[5])
	}
}

func enrichNvidia(gpu *GPUInfo) {
	cmd := exec.Command(
		"nvidia-smi",
		"--query-gpu=memory.total,memory.used,driver_version",
		"--format=csv,noheader,nounits",
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return
	}

	fields := strings.Split(strings.TrimSpace(out.String()), ",")
	if len(fields) >= 3 {
		gpu.VRAMTotalMB = parseUint(fields[0])
		gpu.VRAMUsedMB = parseUint(fields[1])
		gpu.Driver = strings.TrimSpace(fields[2])
	}
}

func parseUint(s string) uint64 {
	s = strings.TrimSpace(s)
	var v uint64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			v = v*10 + uint64(c-'0')
		}
	}
	return v
}
