// Package capture provides platform-specific window discovery and screen capture.
// Use FindWindow to locate a window by title substring, then CaptureWindow to
// grab the pixels in its bounding rectangle.
package capture

import "image"

// WindowInfo holds the title and screen-space bounding rectangle of a window.
type WindowInfo struct {
	// Title is the window title as returned by the platform.
	Title string
	// Bounds is the window's bounding rectangle in screen coordinates.
	Bounds image.Rectangle
}

// FindWindow searches for a window whose title contains titleSubstr (case-sensitive
// substring match). It returns the first matching window or an error when no
// window is found or the lookup fails.
func FindWindow(titleSubstr string) (*WindowInfo, error) {
	return findWindow(titleSubstr)
}

// CaptureWindow captures the screen region described by bounds and returns it as
// an *image.RGBA. The capture is delegated to github.com/kbinani/screenshot so
// the implementation is platform-specific but always CGo-free.
func CaptureWindow(bounds image.Rectangle) (*image.RGBA, error) {
	return captureWindow(bounds)
}
