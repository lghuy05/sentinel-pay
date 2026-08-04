package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
}

func New(addr string, handler http.Handler, shutdownTimeout time.Duration) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
		shutdownTimeout: shutdownTimeout,
	}
}

func (s *Server) ListenAndServe() error {
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	if _, ok := ctx.Deadline(); ok {
		return s.httpServer.Shutdown(ctx)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()

	return s.httpServer.Shutdown(shutdownCtx)
}
