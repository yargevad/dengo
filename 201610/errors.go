package main

import (
	"fmt"
	"mime"
	"net/http"

	"github.com/pkg/errors"
	"github.com/uber-go/zap"
)

type Error struct {
	Code    int
	Message error
}

func (e *Error) Write(w http.ResponseWriter, r *http.Request) {
	// scan request for incoming content-type, to determine response type
	// TODO: middleware to stash media type in context
	//       instead of pulling it from headers and parsing it in multiple places
	ctypeOut := "text/html"
	ctypeIn := r.Header.Get("Content-Type")
	mtype, _, err := mime.ParseMediaType(ctypeIn)
	if err != nil {
		traceErr := errors.New("ParseMediaType failed")
		// structured logs, including stack trace
		env.Log.Error(err.Error(),
			zap.String("content-type", ctypeIn),
			zap.String("trace", fmt.Sprintf("%+v", traceErr)))
	}
	env.Log.Info(errors.Errorf("[%s] (%s) => [%s]", ctypeIn, mtype, ctypeOut).Error())

	// application/json => ditto
	// text/*, none => text/html
}
