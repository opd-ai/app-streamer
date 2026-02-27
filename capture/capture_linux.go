//go:build linux

package capture

import (
	"fmt"
	"image"
	"strings"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/kbinani/screenshot"
)

// findWindow enumerates all X11 managed windows using the _NET_CLIENT_LIST root
// property and returns the first one whose WM_NAME or _NET_WM_NAME contains
// titleSubstr.
func findWindow(titleSubstr string) (*WindowInfo, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, fmt.Errorf("capture: connect to X11: %w", err)
	}
	defer conn.Close()

	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root

	// Intern the atoms we need.
	netClientList, err := internAtom(conn, "_NET_CLIENT_LIST")
	if err != nil {
		return nil, fmt.Errorf("capture: intern _NET_CLIENT_LIST: %w", err)
	}
	netWMName, err := internAtom(conn, "_NET_WM_NAME")
	if err != nil {
		return nil, fmt.Errorf("capture: intern _NET_WM_NAME: %w", err)
	}
	utf8String, err := internAtom(conn, "UTF8_STRING")
	if err != nil {
		return nil, fmt.Errorf("capture: intern UTF8_STRING: %w", err)
	}

	// Retrieve the list of client windows from the root window.
	prop, err := xproto.GetProperty(conn, false, root, netClientList,
		xproto.AtomWindow, 0, 1024).Reply()
	if err != nil {
		return nil, fmt.Errorf("capture: GetProperty _NET_CLIENT_LIST: %w", err)
	}

	wins := prop.Value
	if len(wins) == 0 {
		return nil, fmt.Errorf("capture: no managed windows found")
	}

	// Each window ID is a 4-byte little-endian uint32.
	for i := 0; i+3 < len(wins); i += 4 {
		wid := xproto.Window(uint32(wins[i]) | uint32(wins[i+1])<<8 |
			uint32(wins[i+2])<<16 | uint32(wins[i+3])<<24)

		title, err := getWindowTitle(conn, wid, netWMName, utf8String)
		if err != nil {
			// Fall back to classic WM_NAME.
			title, _ = getWindowTitleClassic(conn, wid)
		}

		if strings.Contains(title, titleSubstr) {
			bounds, err := windowBounds(conn, wid, root)
			if err != nil {
				return nil, fmt.Errorf("capture: window bounds for %q: %w", title, err)
			}
			return &WindowInfo{Title: title, Bounds: bounds}, nil
		}
	}

	return nil, fmt.Errorf("capture: no window with title containing %q", titleSubstr)
}

// captureWindow grabs the pixels of bounds from the screen.
func captureWindow(bounds image.Rectangle) (*image.RGBA, error) {
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, fmt.Errorf("capture: CaptureRect: %w", err)
	}
	return img, nil
}

// internAtom resolves an X atom by name, returning an error on failure.
func internAtom(conn *xgb.Conn, name string) (xproto.Atom, error) {
	reply, err := xproto.InternAtom(conn, true, uint16(len(name)), name).Reply()
	if err != nil {
		return 0, err
	}
	return reply.Atom, nil
}

// getWindowTitle reads the _NET_WM_NAME (UTF-8) property of win.
func getWindowTitle(conn *xgb.Conn, win xproto.Window, nameAtom, utf8Atom xproto.Atom) (string, error) {
	reply, err := xproto.GetProperty(conn, false, win, nameAtom, utf8Atom, 0, 256).Reply()
	if err != nil {
		return "", err
	}
	if reply.ValueLen == 0 {
		return "", fmt.Errorf("empty title")
	}
	return string(reply.Value), nil
}

// getWindowTitleClassic reads the legacy WM_NAME (STRING) property of win.
func getWindowTitleClassic(conn *xgb.Conn, win xproto.Window) (string, error) {
	reply, err := xproto.GetProperty(conn, false, win, xproto.AtomWmName,
		xproto.AtomString, 0, 256).Reply()
	if err != nil {
		return "", err
	}
	return string(reply.Value), nil
}

// windowBounds translates the window's geometry into root (screen) coordinates.
func windowBounds(conn *xgb.Conn, win xproto.Window, root xproto.Window) (image.Rectangle, error) {
	geom, err := xproto.GetGeometry(conn, xproto.Drawable(win)).Reply()
	if err != nil {
		return image.Rectangle{}, fmt.Errorf("GetGeometry: %w", err)
	}

	// TranslateCoordinates maps (0,0) in win's coordinate space to root coords.
	xlate, err := xproto.TranslateCoordinates(conn, win, root, 0, 0).Reply()
	if err != nil {
		return image.Rectangle{}, fmt.Errorf("TranslateCoordinates: %w", err)
	}

	x := int(xlate.DstX)
	y := int(xlate.DstY)
	w := int(geom.Width)
	h := int(geom.Height)
	return image.Rect(x, y, x+w, y+h), nil
}
