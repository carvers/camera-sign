package device

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
)

func ListWebcams(ctx context.Context) ([]Device, error) {
	files, err := filepath.Glob("/dev/video*")
	if err != nil {
		return nil, fmt.Errorf("error listing devices: %w", err)
	}
	devices := make([]Device, 0, len(files))
	for _, file := range files {
		devName := path.Base(file)
		namePath := filepath.Join("/sys/class/video4linux/", devName, "name")
		b, err := ioutil.ReadFile(namePath)
		if err != nil {
			return nil, fmt.Errorf("error getting name of %s: %w", devName, err)
		}
		devices = append(devices, Device{
			Name: strings.TrimSpace(string(b)),
			Path: file,
		})
	}
	return devices, nil
}
