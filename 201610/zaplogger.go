package main

import (
	"net/http"
	"time"

	"github.com/uber-go/zap"
	"github.com/yargevad/chi/middleware"
)

var zapLogFormatter *ZapLogFormatter

func ZapRecoverer(next http.Handler) http.Handler {
	return middleware.FormattedRecoverer(zapLogFormatter, next)
}

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
