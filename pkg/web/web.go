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
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/metrics"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	recoveryBufferSize   = 2048
	serverReadTimeout    = 5 * time.Second
	serverRequestTimeout = 60 * time.Second
	serverWriteTimeout   = 70 * time.Second
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

func checkForServerPanic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, recoveryBufferSize)
				n := runtime.Stack(buf, false)
				buf = buf[:n]

				log.Errorf("recovering from err %v\n %s", err, buf)
				http.Error(w, fmt.Sprintf("server got panic: %v", err), http.StatusInternalServerError)
			}
		}()

		h.ServeHTTP(w, r)
	})
}

func GetHandler() *mux.Router {
	mux := mux.NewRouter()

	mux.Use(checkForServerPanic)
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

var backgroudContext context.Context

func StartServer(ctx context.Context) {
	log.Info(fmt.Sprintf("Starting on %s...", *config.Get().WebListen))

	backgroudContext = ctx

	timeoutMessage := fmt.Sprintf("Server timeout after %s", serverRequestTimeout)

	server := &http.Server{
		Addr:         *config.Get().WebListen,
		Handler:      http.TimeoutHandler(GetHandler(), serverRequestTimeout, timeoutMessage),
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
	}

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), *config.Get().GracefulShutdownTimeout)
		defer cancel()

		_ = server.Shutdown(ctx) //nolint:contextcheck
	}()

	if err := server.ListenAndServe(); err != nil {
		log.WithError(err).Fatal()
	}
}
