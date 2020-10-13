package device

import "errors"

var ErrDeviceListNotSupported = errors.New("device listing is not supported")

type Device struct {
	Name string
	Path string
}
