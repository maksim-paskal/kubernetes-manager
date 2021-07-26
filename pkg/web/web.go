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
	log "github.com/sirupsen/logrus"
)

func GetHandler() *http.ServeMux {
	fs := http.FileServer(http.Dir(*config.Get().FrontDist))

	mux := http.NewServeMux()
	mux.Handle("/", fs)
	mux.HandleFunc("/_nuxt/", serveFiles)
	mux.HandleFunc("/api/getIngress", getIngress)
	mux.HandleFunc("/api/getNamespace", getNamespace)
	mux.HandleFunc("/api/deleteNamespace", deleteNamespace)
	mux.HandleFunc("/api/deleteRegistryTag", deleteRegistryTag)
	mux.HandleFunc("/api/deletePod", deletePod)
	mux.HandleFunc("/api/exec", execCommands)
	mux.HandleFunc("/api/deleteALL", deleteALL)
	mux.HandleFunc("/api/executeBatch", executeBatch)
	mux.HandleFunc("/getKubeConfig", getKubeConfig)
	mux.HandleFunc("/api/scaleNamespace", scaleNamespace)
	mux.HandleFunc("/api/getRunningPodsCount", getRunningPodsCount)
	mux.HandleFunc("/api/version", getAPIversion)
	mux.HandleFunc("/api/getPods", getPods)
	mux.HandleFunc("/api/debug", getDebug)
	mux.HandleFunc("/api/disableHPA", disableHPA)
	// pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return mux
}

func StartServer() {
	log.Info(fmt.Sprintf("Starting on port %d...", *config.Get().Port))

	err := http.ListenAndServe(fmt.Sprintf(":%d", *config.Get().Port), GetHandler())
	if err != nil {
		log.WithError(err).Fatal()
	}
}
