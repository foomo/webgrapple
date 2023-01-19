package httputils

import (
	"context"
	"fmt"
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

func GracefulHttpServer(ctx context.Context, l log.Logger, address string, handler http.Handler) *http.Server {
	httpServer := &http.Server{
		Addr:    address,
		Handler: handler,
	}

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

		shutdownCtx, _ := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			l.Info(fmt.Sprintf("HTTP Server Shutdown Error: %v", err))
		}
		close(idleConnectionsClosed)
	}()
	return httpServer
}
