package device

import (
	"context"
	"fmt"
	"os/exec"
)

func InUse(ctx context.Context, devicePath string, processes ...string) (bool, error) {
	_, err := exec.CommandContext(ctx, "lsof", "-w", "-F", "pcuftn", devicePath).Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 1 {
			if len(e.Stderr) > 0 {
				fmt.Println("lsof stderr: " + string(e.Stderr))
			}
			// an exit code of 1 means device isn't in use
			return false, nil
		}
		return false, fmt.Errorf("error checking %s: %w", devicePath, err)
	}
	return true, nil
}
