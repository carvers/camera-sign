package device

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func InUse(ctx context.Context, devicePath string, processes ...string) (bool, error) {
	if len(processes) == 0 {
		processes = append(processes, "")
	}
	for _, process := range processes {
		args := []string{"-a"}
		if process != "" {
			args = append(args, "-p", process)
		}
		args = append(args, devicePath)
		out, err := exec.CommandContext(ctx, "handle64.exe", args...).Output()
		if err != nil {
			if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 1 {
				if len(e.Stderr) > 0 {
					fmt.Println("handle64 stderr: " + string(e.Stderr))
				}
				// an exit code of 1 means device isn't in use
				continue
			}
			return false, fmt.Errorf("error checking %s: %w", devicePath, err)
		}
		if !strings.Contains(string(out), "No matching handles found.") {
			return true, nil
		}
	}
	return false, nil
}
