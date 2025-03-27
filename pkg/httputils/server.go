package httputils

import (
	"context"
	"fmt"
	syslog "log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/foomo/webgrapple/pkg/log"
)

var (
	DefaultShutdownTimeout = 5 * time.Second
	DefaultShutdownSignals = []os.Signal{
		os.Interrupt,
		os.Kill,
	}
)

type serverErrorLogWriter struct {
	name string
	l    log.Logger
}

func (s *serverErrorLogWriter) Write(p []byte) (int, error) {
	s.l.Error(fmt.Sprintf("%s: %s", s.name, string(p)))
	return len(p), nil
}

func GracefulHTTPServer(ctx context.Context, l log.Logger, name string, address string, handler http.Handler) *http.Server {
	httpServer := &http.Server{
		Addr:    address,
		Handler: handler,
		ErrorLog: syslog.New(&serverErrorLogWriter{
			l:    l,
			name: name,
		}, "", 0),
	}
	cancelCtx := context.WithoutCancel(ctx)

	idleConnectionsClosed := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		for _, s := range DefaultShutdownSignals {
			signal.Notify(sig, s)
		}
		select {
		case s := <-sig:
			l.Info(fmt.Sprintf("server %q shutdown initiated due to signal %v", address, s))
		case <-ctx.Done():
			l.Info(fmt.Sprintf("server %q shutdown initiated due to context", address))
		}

		shutdownCtx, cancel := context.WithTimeout(cancelCtx, DefaultShutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			l.Info(fmt.Sprintf("HTTP Server Shutdown Error: %v", err))
		}
		close(idleConnectionsClosed)
	}()
	return httpServer
}
