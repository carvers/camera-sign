package device

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func InUse(ctx context.Context, devicePath string, processes ...string) (bool, error) {
	out, err := exec.CommandContext(ctx, "is-camera-on").Output()
	if err != nil {
		return false, fmt.Errorf("error checking %s: %w", devicePath, err)
	}
	return strings.TrimSpace(string(out)) == "true", nil
}
