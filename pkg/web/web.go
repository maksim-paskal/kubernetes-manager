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

	"github.com/gorilla/mux"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/metrics"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	errBadFormat     = errors.New("bad format")
	errNoComandFound = errors.New("no command found")
	errMustBePOST    = errors.New("must be POST method")
)

type HandlerResultOutput string

const (
	HandlerResultOutputJSON HandlerResultOutput = "json"
	HandlerResultOutputRAW  HandlerResultOutput = "raw"
)

type HandlerResult struct {
	Version string
	headers map[string]string
	output  HandlerResultOutput
	cached  bool
	Result  interface{}
}

func NewHandlerResult() *HandlerResult {
	return &HandlerResult{
		Version: config.GetVersion(),
		output:  HandlerResultOutputJSON,
		headers: make(map[string]string),
	}
}

func GetHandler() *mux.Router {
	mux := mux.NewRouter()

	mux.HandleFunc("/api/ready", handlerReady)
	mux.HandleFunc("/api/healthz", handlerHealthz)
	mux.HandleFunc("/oauth2/userinfo", handlerUser)
	mux.HandleFunc("/api/{operation}", handlerAPI)
	mux.HandleFunc("/api/{environmentID}/{operation}", handlerEnvironment)

	// pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// metrics
	mux.Handle("/metrics", metrics.GetHandler())

	mux.PathPrefix("/_nuxt").Handler(NewHandlerSPACached(*config.Get().FrontDist, "index.html"))
	mux.PathPrefix("/").Handler(NewHandlerSPA(*config.Get().FrontDist, "index.html"))

	return mux
}

func StartServer() {
	log.Info(fmt.Sprintf("Starting on port %d...", *config.Get().Port))

	err := http.ListenAndServe(fmt.Sprintf(":%d", *config.Get().Port), GetHandler())
	if err != nil {
		log.WithError(err).Fatal()
	}
}
