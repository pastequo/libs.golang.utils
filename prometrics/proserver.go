// Package prometrics contains helper to start/stop a server to allow Prometheus to scrap the metrics.
// It also contains basic metrics to evaluate how often each route are called and their responses.
package prometrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pastequo/libs.golang.utils/logutil"
)

var (
	startOnce sync.Once
	started   chan bool

	stopOnce  sync.Once
	stopError error

	srv *http.Server
)

// ErrNotStarted is returned if Shutdown is called before StartServer.
var ErrNotStarted = errors.New("server not started")

// StartServer starts the prometheus handler and route it to the given port.
func StartServer(port uint) {

	startOnce.Do(func() {
		logger := logutil.GetDefaultLogger()
		logger.Debugf("Starting prometheus server on %v...", port)

		started = make(chan bool, 1)

		srv = &http.Server{Addr: fmt.Sprintf(":%v", port)}
		srv.SetKeepAlivesEnabled(true)
		srv.IdleTimeout = 5 * time.Second

		router := http.NewServeMux()
		router.Handle("/metrics", promhttp.Handler())
		srv.Handler = router

		go func() {
			close(started)
			err := srv.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				logger.WithError(err).Warn("prometheus server stopped")
			}
		}()
	})
}

// Shutdown stops the prometheus server.
func Shutdown(ctx context.Context) error {

	logger := logutil.GetDefaultLogger()
	logger.Debugf("Stopping prometheus server...")

	select {
	case <-started:
	case <-ctx.Done():
		return ErrNotStarted
	}

	stopOnce.Do(func() {
		stopError = srv.Shutdown(ctx)
		if stopError != nil {
			logger.WithError(stopError).Warn("failed to stop prometheus server")
		}
	})

	return stopError
}
