# app-streamer

Stream a desktop application window as live MJPEG video over HTTP — pure Go,
zero CGo, works in any browser.

## Usage

```
app-streamer [flags] <window-title>
```

Open `http://localhost:8080/` in a browser to see the live stream.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `:8080` | HTTP listen address |
| `-fps` | `10` | Capture frames per second |
| `-quality` | `75` | JPEG quality (1–100) |

### Examples

```sh
# Stream the Firefox window
app-streamer "Firefox"

# Stream on a custom port at 30 FPS
app-streamer -addr :9000 -fps 30 "Terminal"
```

## Build

```sh
# Linux
CGO_ENABLED=0 go build -o app-streamer .

# Windows (cross-compile from Linux)
CGO_ENABLED=0 GOOS=windows go build -o app-streamer.exe .
```

## Platform Support

| Platform | Window Lookup | Screen Capture |
|----------|---------------|----------------|
| Linux (X11) | `jezek/xgb` (pure Go XCB) | `kbinani/screenshot` |
| Windows | `golang.org/x/sys/windows` syscalls | `kbinani/screenshot` |
| Other | Returns an error at runtime | — |

## Library Attribution

| Library | Use |
|---------|-----|
| [`github.com/kbinani/screenshot`](https://github.com/kbinani/screenshot) | CGo-free screen region capture (Linux X11, Windows, macOS) |
| [`github.com/saljam/mjpeg`](https://github.com/saljam/mjpeg) | MJPEG-over-HTTP streaming |
| [`github.com/jezek/xgb`](https://github.com/jezek/xgb) | Pure-Go X11 protocol (window enumeration on Linux) |
| [`golang.org/x/sys`](https://pkg.go.dev/golang.org/x/sys) | Windows system DLL wrappers |
