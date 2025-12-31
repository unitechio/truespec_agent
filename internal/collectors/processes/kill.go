package processes

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

func Kill(ctx context.Context, pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	running, _ := p.IsRunning()
	if !running {
		return nil
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		_ = p.SendSignal(syscall.SIGTERM)
		time.Sleep(2 * time.Second)
		return p.Kill()

	case "windows":
		return p.Kill()

	default:
		return fmt.Errorf("unsupported os")
	}
}
