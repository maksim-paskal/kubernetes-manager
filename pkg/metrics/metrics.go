/*
Copyright paskal.maksim@gmail.com
Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package metrics

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "kubernetes_manager"

type Instrumenter struct {
	subsystemIdentifier        string
	inFlightRequestsGauge      prometheus.Gauge
	requestsPerEndpointCounter *prometheus.CounterVec
	requestLatencyHistogram    *prometheus.HistogramVec
	PathDetectionFunc          func(r *http.Request) string
}

// New creates a new Instrumenter. The subsystemIdentifier will be used as part of
// the metric names (e.g. http_<identifier>_requests_total).
func NewInstrumenter(subsystemIdentifier string) *Instrumenter {
	name := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(subsystemIdentifier, "")

	return &Instrumenter{
		subsystemIdentifier: name,
		PathDetectionFunc: func(r *http.Request) string {
			return r.URL.Path
		},
		inFlightRequestsGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      fmt.Sprintf("http_%s_in_flight_requests", name),
			Help:      fmt.Sprintf("A gauge of in-flight requests to the http %s.", name),
		}),

		requestsPerEndpointCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      fmt.Sprintf("http_%s_requests_total", name),
				Help:      fmt.Sprintf("A counter for requests to the http %s per endpoint.", name),
			},
			[]string{"code", "method", "endpoint"},
		),

		requestLatencyHistogram: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      fmt.Sprintf("http_%s_request_duration_seconds", name),
				Help:      fmt.Sprintf("A histogram of request latencies to the http %s .", name),
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method"},
		),
	}
}

// InstrumentedRoundTripper returns an instrumented round tripper.
func (i *Instrumenter) InstrumentedRoundTripper() http.RoundTripper {
	return i.InstrumentedRoundTripperWithTransport(http.DefaultTransport)
}

func (i *Instrumenter) InstrumentedRoundTripperWithTransport(defaultTransport http.RoundTripper) http.RoundTripper {
	return promhttp.InstrumentRoundTripperInFlight(i.inFlightRequestsGauge,
		promhttp.InstrumentRoundTripperDuration(i.requestLatencyHistogram,
			i.instrumentRoundTripperEndpoint(i.requestsPerEndpointCounter,
				defaultTransport,
			),
		),
	)
}

func (i *Instrumenter) instrumentRoundTripperEndpoint(counter *prometheus.CounterVec, next http.RoundTripper) promhttp.RoundTripperFunc { //nolint:lll
	return func(r *http.Request) (*http.Response, error) {
		_, span := telemetry.Start(r.Context(), "http."+i.subsystemIdentifier)
		defer span.End()

		spanAttributes := map[string]string{
			"url":    r.URL.String(),
			"method": r.Method,
		}

		resp, err := next.RoundTrip(r)
		if err == nil {
			i.incrementCounterWithRequest(counter, resp)
		}

		if resp != nil {
			spanAttributes["status"] = resp.Status
		}

		telemetry.Attributes(span, spanAttributes)

		return resp, errors.Wrap(err, "error making roundtrip")
	}
}

func (i *Instrumenter) incrementCounterWithRequest(counter *prometheus.CounterVec, resp *http.Response) {
	if resp == nil {
		return
	}

	if resp.Request == nil {
		return
	}

	if resp.Request.URL == nil {
		return
	}

	statusCode := strconv.Itoa(resp.StatusCode)
	method := strings.ToLower(resp.Request.Method)
	path := i.PathDetectionFunc(resp.Request)

	counter.WithLabelValues(statusCode, method, path).Inc()
}

var (
	KubernetesAPIRequest = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "apiserver_request_total",
		Help:      "The total number of kunernetes API requests",
	}, []string{"cluster", "code"})

	KubernetesAPIRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "apiserver_request_duration",
		Help:      "The duration in seconds of kunernetes API requests",
	}, []string{"cluster"})

	WebRequest = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "request_total",
		Help:      "The total number of web requests",
	}, []string{"operation"})

	WebRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "request_duration",
		Help:      "The duration in seconds of web requests",
	}, []string{"operation"})

	CacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "cache_hits_total",
		Help:      "The total number of cache hits",
	}, []string{"operation"})

	CacheAdd = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "cache_add_total",
		Help:      "The total number of cache adds",
	})

	CacheRemoved = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "cache_remove_total",
		Help:      "The total number of cache removes",
	})
)

func LogRequest(operation string, startTime time.Time) {
	WebRequest.WithLabelValues(operation).Inc()
	WebRequestDuration.WithLabelValues(operation).Observe(time.Since(startTime).Seconds())
}

func GetHandler() http.Handler {
	return promhttp.Handler()
}
