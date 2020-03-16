package main

import (
	"net/http"

	sentry "github.com/getsentry/sentry-go"
	opentracing "github.com/opentracing/opentracing-go"
)

func logError(span opentracing.Span, level sentry.Level, request *http.Request, err error, message string) {
	span.SetTag("error", true)

	localHub := sentry.CurrentHub().Clone()
	localHub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
		if request != nil {
			scope.SetExtra("Request.Header", request.Header)
			scope.SetExtra("Request.Cookies", request.Cookies())
			scope.SetExtra("Request.RemoteAddr", request.RemoteAddr)
			scope.SetExtra("Request.URL", request.URL)
			scope.SetExtra("Request.URL.Query", request.URL.Query())
			scope.SetExtra("Request.PostForm", request.PostForm)
		}
	})
	if err != nil {
		localHub.CaptureException(err)
		span.LogKV("error", err)
	} else {
		localHub.CaptureMessage(message)
		span.LogKV("error", message)
	}
}
