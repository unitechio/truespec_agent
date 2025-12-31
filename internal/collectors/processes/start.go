package processes

import (
	"context"
	"fmt"
	"os/exec"
)

func Start(ctx context.Context, path string, args []string) (int32, error) {
	cmd := exec.CommandContext(ctx, path, args...)

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start failed: %w", err)
	}

	return int32(cmd.Process.Pid), nil
}
