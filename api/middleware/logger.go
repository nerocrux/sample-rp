package middleware

import (
	"fmt"
	"net/http"

	"github.com/morikuni/failure"
	"github.com/nerocrux/sample-rp/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type HttpHandler func(w http.ResponseWriter, r *http.Request) error

func HTTPLogger(h HttpHandler, l *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		if !errors.IsFailureError(err) {
			l.Error(err.Error())
			return
		}

		callstack, _ := failure.CallStackOf(err)
		fields := []zapcore.Field{
			zap.String("stacktrace", fmt.Sprintf("%v", callstack.Frames())),
		}

		code, _ := failure.CodeOf(err)
		switch code {
		case errors.InvalidArguments, errors.NotFoundEntity:
			l.Warn(err.Error(), fields...)
		case errors.Internal:
			l.Error(err.Error(), fields...)
		default:
			l.Error(err.Error(), fields...)
		}
	})
}
