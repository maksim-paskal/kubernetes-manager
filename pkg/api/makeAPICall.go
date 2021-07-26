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
package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
)

type HTTPResponse struct {
	Status string
	Body   string
}

func MakeAPICall(span opentracing.Span, api string, q url.Values, ch chan<- HTTPResponse) {
	url := fmt.Sprintf("http://%s:%d%s", *config.Get().MakeAPICallServer, *config.Get().Port, api)

	ctx := context.Background()

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.URL.RawQuery = q.Encode()

	tracer := opentracing.GlobalTracer()
	err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	if err != nil { //nolint:wsl
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			Error()
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			Error()
	} else if resp.Body != nil {
		defer resp.Body.Close()
	}

	httpBody, _ := ioutil.ReadAll(resp.Body)

	ch <- HTTPResponse{resp.Status, string(httpBody)}
}
