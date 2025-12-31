package processes

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/process"
)

func Restart(ctx context.Context, pid int32) (int32, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return 0, err
	}

	cmdline, err := p.CmdlineSlice()
	if err != nil || len(cmdline) == 0 {
		return 0, fmt.Errorf("cannot get cmdline")
	}

	exe, err := p.Exe()
	if err != nil {
		return 0, err
	}

	_ = Kill(ctx, pid)

	return Start(ctx, exe, cmdline[1:])
}
