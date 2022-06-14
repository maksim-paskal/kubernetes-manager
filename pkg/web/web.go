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
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/metrics"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func GetHandler() *http.ServeMux {
	fs := http.FileServer(http.Dir(*config.Get().FrontDist))

	mux := http.NewServeMux()
	mux.Handle("/", fs)
	mux.HandleFunc("/_nuxt/", serveFiles)
	mux.HandleFunc("/api/getIngress", getIngress)
	mux.HandleFunc("/api/deleteNamespace", deleteNamespace)
	mux.HandleFunc("/api/deleteRegistryTag", deleteRegistryTag)
	mux.HandleFunc("/api/deletePod", deletePod)
	mux.HandleFunc("/api/exec", execCommands)
	mux.HandleFunc("/api/deleteALL", deleteALL)
	mux.HandleFunc("/api/executeBatch", executeBatch)
	mux.HandleFunc("/getKubeConfig", getKubeConfig)
	mux.HandleFunc("/api/scaleNamespace", scaleNamespace)
	mux.HandleFunc("/api/scaleDownDelay", scaleDownDelay)
	mux.HandleFunc("/api/getRunningPodsCount", getRunningPodsCount)
	mux.HandleFunc("/api/version", getAPIversion)
	mux.HandleFunc("/api/getPods", getPods)
	mux.HandleFunc("/api/debug", getDebug)
	mux.HandleFunc("/api/disableHPA", disableHPA)
	mux.HandleFunc("/api/disableMTLS", disableMTLS)
	mux.HandleFunc("/api/getProjects", getProjects)
	mux.HandleFunc("/api/getProjectRefs", getProjectRefs)
	mux.HandleFunc("/api/getProjectInfo", getProjectInfo)
	mux.HandleFunc("/api/deploySelectedServices", deploySelectedServices)
	mux.HandleFunc("/api/createNewBranch", createNewBranch)
	mux.HandleFunc("/api/getServices", getServices)
	mux.HandleFunc("/api/getFrontConfig", getFrontConfig)
	mux.HandleFunc("/api/ready", handlerReady)
	mux.HandleFunc("/api/healthz", handlerHealthz)

	// pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// metrics
	mux.Handle("/metrics", metrics.GetHandler())

	return mux
}

func StartServer() {
	log.Info(fmt.Sprintf("Starting on port %d...", *config.Get().Port))

	err := http.ListenAndServe(fmt.Sprintf(":%d", *config.Get().Port), GetHandler())
	if err != nil {
		log.WithError(err).Fatal()
	}
}

func checkParams(r *http.Request, params []string) error {
	for _, param := range params {
		if len(r.URL.Query()[param]) != 1 {
			return errors.Wrap(errNoQueryParam, param)
		}
	}

	return nil
}

func handlerReady(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("ready"))
	if err != nil {
		if err != nil {
			log.
				WithError(err).
				WithFields(logrushooksentry.AddRequest(r)).
				Error()
		}
	}
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("live"))
	if err != nil {
		if err != nil {
			log.
				WithError(err).
				WithFields(logrushooksentry.AddRequest(r)).
				Error()
		}
	}
}
