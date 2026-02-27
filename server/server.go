// Package server provides the MJPEG HTTP server for app-streamer.
// It exposes two endpoints:
//
//   - GET /        – an HTML page that embeds the live MJPEG stream in an <img> tag
//   - GET /stream  – the raw MJPEG stream consumed by browsers and media players
package server

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"log/slog"
	"net/http"

	"github.com/saljam/mjpeg"
)

// Server wraps an MJPEG stream and an HTTP server.
type Server struct {
	stream *mjpeg.Stream
	srv    *http.Server
}

// indexHTML is the minimal HTML page served at GET /.
const indexHTML = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>app-streamer</title></head>
<body style="margin:0;background:#000;display:flex;justify-content:center;align-items:center;min-height:100vh">
  <img src="/stream" style="max-width:100%;max-height:100vh" alt="live stream">
</body>
</html>
`

// New creates a Server that will listen on addr.
func New(addr string) *Server {
	stream := mjpeg.NewStream()

	mux := http.NewServeMux()
	mux.Handle("/stream", stream)
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(indexHTML))
	})

	return &Server{
		stream: stream,
		srv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

// UpdateFrame encodes img as JPEG at the given quality and pushes it to all
// connected MJPEG clients. quality must be in [1, 100].
func (s *Server) UpdateFrame(img *image.RGBA, quality int) error {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return err
	}
	s.stream.UpdateJPEG(buf.Bytes())
	return nil
}

// Start begins listening and serving HTTP requests. It blocks until the server
// is shut down and returns http.ErrServerClosed on a clean shutdown.
func (s *Server) Start() error {
	addr := s.srv.Addr
	// When addr is ":port", prepend localhost for the log message.
	displayAddr := addr
	if len(addr) > 0 && addr[0] == ':' {
		displayAddr = "localhost" + addr
	}
	slog.Info("HTTP server listening", "addr", addr)
	slog.Info("Open http://"+displayAddr+"/ in your browser")
	return s.srv.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server using the provided context.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
