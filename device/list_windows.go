package device

import (
	"context"
)

func ListWebcams(ctx context.Context) ([]Device, error) {
	return nil, ErrDeviceListNotSupported
}
