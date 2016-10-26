package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pressly/chi/middleware"
	"github.com/uber-go/zap"
)

type Error struct {
	Code    int
	Message error
}

func (e *Error) Write(w http.ResponseWriter, r *http.Request) {
	var ctype string
	var ok bool
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)

	// Default Content-Type is HTML if it's somehow not set
	if ctype, ok = ctx.Value("responseType").(string); !ok {
		ctype = "text/html"
	}

	// Write error to local logs.
	env.Log.Error("error",
		zap.Error(e.Message),
		zap.Int("code", e.Code),
		zap.String("reqID", reqID))

	w.Header().Set("Content-Type", ctype)
	w.WriteHeader(e.Code)

	if ctype == "application/json" {
		// Write JSON-encoded error to client
		if err := json.NewEncoder(w).Encode(e); err != nil {
			// XXX: If we can't write out an error, log and continue
			env.Log.Error("error encode failed",
				zap.Error(err),
				zap.String("reqID", reqID))
		}

	} else if ctype == "text/html" {
		// FIXME: use an actual template here
		fmt.Fprintf(w, "<html><head><title>%d error</title></head><body><h1>%d</h1><h4>%s</h4</body></html>",
			e.Code, e.Code, e.Message)
	}
}
