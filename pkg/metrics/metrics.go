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
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "kubernetes_manager"

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
)

func GetHandler() http.Handler {
	return promhttp.Handler()
}
