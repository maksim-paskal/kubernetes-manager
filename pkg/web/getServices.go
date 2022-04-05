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
package web

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log "github.com/sirupsen/logrus"
)

const mysqlRegexpFilter = ".*mysql.*\\.svc\\.cluster\\.local$"

func getServices(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getServices", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	if err := checkParams(r, []string{"namespace"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	namespace := r.URL.Query()["namespace"]

	services, err := api.GetServices(namespace[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	type ResultType struct {
		Result []*api.GetServicesItem `json:"result"`
	}

	filterRegexp := regexp.MustCompile(mysqlRegexpFilter)

	mysqlServices := make([]*api.GetServicesItem, 0)
	otherServices := make([]*api.GetServicesItem, 0)

	for _, service := range services {
		if filterRegexp.MatchString(service.ServiceHost) {
			mysqlServices = append(mysqlServices, service)
		} else if !strings.Contains(service.ServiceHost, ".svc.cluster.local") {
			// display only pods
			otherServices = append(otherServices, service)
		}
	}

	result := ResultType{
		Result: append(mysqlServices, otherServices...),
	}

	js, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)

	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()
	}
}
