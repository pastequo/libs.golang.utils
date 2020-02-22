package prometrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type metricResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newMetricResponseWriter(w http.ResponseWriter) *metricResponseWriter {
	return &metricResponseWriter{w, http.StatusOK}
}

func (lrw *metricResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// GetMiddleware provides a middleware that monitors how often each route is called and its response.
func GetMiddleware(namespace, appName string, handler http.Handler) (http.Handler, error) {

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "endpoint_request_count",
		Help:      "collage request count.",
	}, []string{"app", "name", "method", "state"})

	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "endpoint_duration_milliseconds",
		Help:      "Time taken to execute endpoint.",
	}, []string{"app", "name", "method", "status"})

	err := prometheus.Register(histogram)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to register %v %v histogram", namespace, appName)
	}

	err = prometheus.Register(counter)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to register %v %v counter", namespace, appName)
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := newMetricResponseWriter(w)

		handler.ServeHTTP(lrw, r)

		// hack for healthcheck
		if r.URL.Path == "/healthcheck" {
			return
		}

		statusCode := lrw.statusCode
		duration := time.Since(start)
		durationMilli := float64(duration/time.Millisecond) + float64(duration%time.Millisecond)/float64(time.Millisecond)

		histogram.WithLabelValues(appName, r.URL.Path, r.Method, fmt.Sprintf("%d", statusCode)).Observe(durationMilli)
		counter.WithLabelValues(appName, r.URL.Path, r.Method, fmt.Sprintf("%d", statusCode)).Inc()
	})

	return h, nil
}
