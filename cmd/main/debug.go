package main

import (
	"fmt"
	"net/http"
	"strings"

	sentry "github.com/getsentry/sentry-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func getDebug(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getDebug", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	_, err := w.Write(formatRequest(span, r))

	if err != nil {
		logError(span, sentry.LevelError, r, err, "")
	}
}

func formatRequest(span opentracing.Span, r *http.Request) []byte {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			logError(span, sentry.LevelError, r, err, "")
		}
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	return []byte(strings.Join(request, "\n"))
}
