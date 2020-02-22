// Package panicrecover contains a panic handler middleware.
package panicrecover

import (
	"net/http"
	"runtime/debug"

	"github.com/pastequo/libs.golang.utils/logutil"
)

// GetMiddleware provides a middleware that recovers from crash.
func GetMiddleware(handler http.Handler) http.Handler {

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		defer func() {
			r := recover()
			if r != nil {

				logger := logutil.GetLogger(req.Context())
				logger.Errorf("Recovering from panic: %v, from request %v", r, *req)
				debug.PrintStack()

				resp.WriteHeader(http.StatusInternalServerError)
			}
		}()

		handler.ServeHTTP(resp, req)
	})
}
