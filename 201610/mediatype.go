package main

import (
	"context"
	"mime"
	"net/http"
	"strings"

	"github.com/pressly/chi/middleware"
	"github.com/uber-go/zap"
)

// SetOutputType sets the output type based on the incoming Content-Type header.
func SetOutputType(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Default response Content-Type is HTML
		responseType := "text/html"
		contentType := r.Header.Get("Content-Type")
		if contentType != "" {
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				// XXX: If we can't parse the request Content-Type, log and continue
				env.Log.Warn("ParseMediaType failed",
					zap.String("content-type", contentType),
					zap.String("reqID", middleware.GetReqID(ctx)))
			} else if mediaType == "application/json" || strings.HasPrefix(mediaType, "form/") {
				responseType = mediaType
			}
		}
		r = r.WithContext(context.WithValue(ctx, "responseType", responseType))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
