// Package queryid contains a middleware handler that prepares the context for future loggers.
package queryid

import (
	"net/http"

	"github.com/pastequo/libs.golang.utils/logutil"
)

// GetMiddleware provides a middleware that add the queryID to the ctx.
func GetMiddleware(handler http.Handler) http.Handler {

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {

		path := ""
		verb := ""
		if req != nil {
			verb = req.Method
			if req.URL != nil {
				path = req.URL.Path
			}
		}

		ctx := logutil.UpdateContext(req.Context(), "query", verb, path)
		handler.ServeHTTP(resp, req.WithContext(ctx))
	})
}
