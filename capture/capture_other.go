//go:build !linux && !windows

package capture

import (
	"fmt"
	"image"
	"runtime"
)

func findWindow(_ string) (*WindowInfo, error) {
	return nil, fmt.Errorf("capture: unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
}

func captureWindow(_ image.Rectangle) (*image.RGBA, error) {
	return nil, fmt.Errorf("capture: unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
}
