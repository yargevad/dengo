package main

import (
	"context"
	"mime"
	"net/http"

	"github.com/pkg/errors"
)

func RequestType(r *http.Request) (mediaType string, err error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" {
		mediaType, _, err = mime.ParseMediaType(contentType)
		if err != nil {
			return mediaType, errors.Wrap(err, "ParseMediaType failed")
		}
	}
	return mediaType, nil
}

func ContentTypeChecks(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var inType string
		var err error
		var e *Error

		if r.Method == "POST" {
			if inType, err = RequestType(r); err != nil {
				e = &Error{Code: http.StatusBadRequest, Message: err}
				e.Write(w, r)
				return
			}
			// content-type is required
			if inType == "" {
				e = &Error{
					Code:    http.StatusBadRequest,
					Message: errors.New("POST requests require a content-type"),
				}
				e.Write(w, r)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), "content-type", inType))
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
