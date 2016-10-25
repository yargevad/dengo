package main

import (
	"mime"
	"net/http"
)

func RequestType(r *http.Request) (mediaType string, err error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" {
		mediaType, _, err = mime.ParseMediaType(contentType)
	}
	return mediaType, err
}
