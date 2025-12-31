package processes

import (
	"context"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"
)

func Stop(ctx context.Context, pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}

	return p.SendSignal(syscall.SIGTERM)
}
