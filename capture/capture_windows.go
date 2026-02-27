//go:build windows

package capture

import (
	"fmt"
	"image"
	"strings"
	"syscall"
	"unsafe"

	"github.com/kbinani/screenshot"
	"golang.org/x/sys/windows"
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procEnumWindows         = user32.NewProc("EnumWindows")
	procGetWindowText       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLength = user32.NewProc("GetWindowTextLengthW")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procIsWindowVisible     = user32.NewProc("IsWindowVisible")
)

// enumState is used by the package-level EnumWindows callback to communicate
// the search criteria and the result without needing unsafe lParam casts.
// findWindow serialises access so no mutex is needed.
var enumState struct {
	titleSubstr string
	found       windows.HWND
	title       string
}

// enumWindowsProc is called for each top-level window by EnumWindows.
func enumWindowsProc(hwnd windows.HWND, _ uintptr) uintptr {
	// Skip invisible windows.
	visible, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
	if visible == 0 {
		return 1 // continue
	}

	length, _, _ := procGetWindowTextLength.Call(uintptr(hwnd))
	if length == 0 {
		return 1
	}

	buf := make([]uint16, length+1)
	procGetWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), length+1)
	title := syscall.UTF16ToString(buf)

	if strings.Contains(title, enumState.titleSubstr) {
		enumState.found = hwnd
		enumState.title = title
		return 0 // stop enumeration
	}
	return 1 // continue
}

// findWindow uses EnumWindows to locate a visible top-level window whose title
// contains titleSubstr.
func findWindow(titleSubstr string) (*WindowInfo, error) {
	enumState.titleSubstr = titleSubstr
	enumState.found = 0
	enumState.title = ""

	cb := syscall.NewCallback(enumWindowsProc)
	procEnumWindows.Call(cb, 0)

	if enumState.found == 0 {
		return nil, fmt.Errorf("capture: no window with title containing %q", titleSubstr)
	}

	bounds, err := getWindowRect(enumState.found)
	if err != nil {
		return nil, fmt.Errorf("capture: GetWindowRect for %q: %w", enumState.title, err)
	}
	return &WindowInfo{Title: enumState.title, Bounds: bounds}, nil
}

// captureWindow grabs the pixels of bounds from the screen.
func captureWindow(bounds image.Rectangle) (*image.RGBA, error) {
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, fmt.Errorf("capture: CaptureRect: %w", err)
	}
	return img, nil
}

// RECT mirrors the Win32 RECT structure.
type winRECT struct {
	Left, Top, Right, Bottom int32
}

// getWindowRect returns the screen-space bounding rectangle of hwnd.
func getWindowRect(hwnd windows.HWND) (image.Rectangle, error) {
	var r winRECT
	ret, _, err := procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
	if ret == 0 {
		return image.Rectangle{}, fmt.Errorf("GetWindowRect: %w", err)
	}
	return image.Rect(int(r.Left), int(r.Top), int(r.Right), int(r.Bottom)), nil
}
