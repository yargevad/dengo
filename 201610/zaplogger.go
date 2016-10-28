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

type ZapLogFormatter struct{}

func (z *ZapLogFormatter) FormatLog(r *http.Request, code, nbytes int, elapsed time.Duration, err error) {
	var f10 [10]zap.Field
	var fields []zap.Field = f10[:0]

	reqID := middleware.GetReqID(r.Context())
	if reqID != "" {
		fields = append(fields, zap.String("reqID", reqID))
	}

	fields = append(fields,
		zap.String("method", r.Method),
		zap.String("host", r.Host),
		zap.String("uri", r.RequestURI),
		zap.String("proto", r.Proto),
		zap.String("remote", r.RemoteAddr))

	if r.TLS != nil {
		fields = append(fields, zap.Bool("tls", true))
	} else {
		fields = append(fields, zap.Bool("tls", false))
	}

	if err == nil {
		fields = append(fields, zap.Int("code", code),
			zap.Int("bytes", nbytes),
			zap.Duration("elapsed", elapsed))
	} else {
		fields = append(fields, zap.Error(err), zap.Stack())
	}

	switch {
	case code < 500:
		env.Log.Info("served", fields...)
	default:
		env.Log.Warn("error", fields...)
	}
}
