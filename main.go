// app-streamer captures a desktop application window and streams it as live
// MJPEG video over HTTP. Usage:
//
//	app-streamer [flags] <window-title>
//
// Flags:
//
//	-addr    string   HTTP listen address (default ":8080")
//	-fps     int      Capture frames per second (default 10)
//	-quality int      JPEG quality 1-100 (default 75)
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opd-ai/app-streamer/capture"
	"github.com/opd-ai/app-streamer/server"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	fps := flag.Int("fps", 10, "Capture frames per second")
	quality := flag.Int("quality", 75, "JPEG quality 1-100")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: app-streamer [flags] <window-title>\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	windowTitle := flag.Arg(0)

	if *fps <= 0 {
		fmt.Fprintln(os.Stderr, "app-streamer: -fps must be positive")
		os.Exit(1)
	}
	if *quality < 1 || *quality > 100 {
		fmt.Fprintln(os.Stderr, "app-streamer: -quality must be between 1 and 100")
		os.Exit(1)
	}

	// Locate the target window once to verify it exists.
	slog.Info("Looking for window", "title", windowTitle)
	win, err := capture.FindWindow(windowTitle)
	if err != nil {
		fmt.Fprintf(os.Stderr, "app-streamer: %v\n", err)
		os.Exit(1)
	}
	slog.Info("Found window", "title", win.Title, "bounds", win.Bounds)

	// Set up the MJPEG HTTP server.
	srv := server.New(*addr)

	// Capture loop runs in a background goroutine.
	interval := time.Duration(float64(time.Second) / float64(*fps))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Re-resolve the window on every frame so we track moves/resizes.
				current, err := capture.FindWindow(windowTitle)
				if err != nil {
					slog.Warn("Window not found, retrying", "title", windowTitle, "err", err)
					continue
				}
				img, err := capture.CaptureWindow(current.Bounds)
				if err != nil {
					slog.Warn("Capture failed", "err", err)
					continue
				}
				if err := srv.UpdateFrame(img, *quality); err != nil {
					slog.Warn("Frame encode failed", "err", err)
				}
			}
		}
	}()

	// Handle OS signals for graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		slog.Info("Received signal, shutting down", "signal", sig)
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("HTTP shutdown error", "err", err)
		}
	}()

	// Start the HTTP server (blocks until shutdown).
	if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintf(os.Stderr, "app-streamer: HTTP server: %v\n", err)
		os.Exit(1)
	}
	slog.Info("Bye.")
}
