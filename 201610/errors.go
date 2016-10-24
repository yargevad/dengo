package main

import (
	"fmt"
	"mime"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/uber-go/zap"
	"github.com/yargevad/chi/middleware"
)

type Error struct {
	Code    int
	Message error
}

var zapLogFormatter *ZapLogFormatter

func ZapLogger(next http.Handler) http.Handler {
	return middleware.FormattedLogger(zapLogFormatter, next)
}

type ZapLogFormatter struct {
	code   int
	fields []zap.Field
}

func (z *ZapLogFormatter) FormatRequest(r *http.Request) middleware.LogFormatter {
	var ctx ZapLogFormatter
	ctx.fields = make([]zap.Field, 0, 10)

	reqID := middleware.GetReqID(r.Context())
	if reqID != "" {
		ctx.fields = append(ctx.fields, zap.String("reqID", reqID))
	}

	ctx.fields = append(ctx.fields,
		zap.String("method", r.Method),
		zap.String("host", r.Host),
		zap.String("uri", r.RequestURI),
		zap.String("proto", r.Proto),
		zap.String("remote", r.RemoteAddr))

	if r.TLS != nil {
		ctx.fields = append(ctx.fields, zap.Bool("tls", true))
	} else {
		ctx.fields = append(ctx.fields, zap.Bool("tls", false))
	}

	return &ctx
}

func (ctx *ZapLogFormatter) FormatResponse(code, bytes int, elapsed time.Duration) {
	ctx.code = code
	ctx.fields = append(ctx.fields, zap.Int("code", code))
	ctx.fields = append(ctx.fields, zap.Int("bytes", bytes))
	ctx.fields = append(ctx.fields, zap.Duration("elapsed", elapsed))
}

func (ctx *ZapLogFormatter) Log() {
	switch {
	case ctx.code < 500:
		env.Log.Info("served", ctx.fields...)
	default:
		env.Log.Warn("error", ctx.fields...)
	}
}

func (ctx *ZapLogFormatter) Recover(err error) {
	ctx.fields = append(ctx.fields, zap.Error(err.(error)))
	ctx.Log()
}

func (e *Error) Write(w http.ResponseWriter, r *http.Request) {
	// TODO: middleware to stash media type in context
	//       instead of pulling it from headers and parsing it in multiple places
	var ctypeIn, ctypeOut, mtype string
	var err, trace error

	// scan request for incoming content-type, to determine response type
	ctypeOut = "text/html"
	ctypeIn = r.Header.Get("Content-Type")
	if ctypeIn != "" {
		mtype, _, err = mime.ParseMediaType(ctypeIn)
		if err != nil {
			trace = errors.New("ParseMediaType failed")
			env.Log.Error(err.Error(),
				zap.String("content-type", ctypeIn),
				zap.String("trace", fmt.Sprintf("%+v", trace)))
		}
	}
	env.Log.Info(errors.Errorf("[%s] (%s) => [%s]", ctypeIn, mtype, ctypeOut).Error())

	// application/json => ditto
	// text/*, none => text/html
}
