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
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	log "github.com/sirupsen/logrus"
)

type HandlerSPA struct {
	staticPath string
	indexPath  string
	replacer   *strings.Replacer
	cached     bool
}

func NewHandlerSPA(staticPath, indexPath string) *HandlerSPA {
	h := HandlerSPA{
		staticPath: staticPath,
		indexPath:  indexPath,
		cached:     false,
	}

	return &h
}

func NewHandlerSPACached(staticPath, indexPath string) *HandlerSPA {
	h := HandlerSPA{
		staticPath: staticPath,
		indexPath:  indexPath,
		cached:     true,
	}

	h.replacer = h.GetContentReplacer()

	return &h
}

func (h HandlerSPA) GetContentReplacer() *strings.Replacer {
	return strings.NewReplacer(
		"https://__setry_id__@__setry_server__/1", config.GetEnvDefault("FRONT_SENTRY_DSN", "https://id@sentry/1"),
	)
}

func (h HandlerSPA) serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(*config.Get().FrontDist, filepath.Clean(r.URL.Path))

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		http.Error(w, err.Error(), http.StatusNotFound)

		return
	}

	read, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	newContents := h.replacer.Replace(string(read))

	w.Header().Set("Cache-Control", "public, max-age=86400")

	_, err = w.Write([]byte(newContents))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}
}

func (h HandlerSPA) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.cached {
		h.serveStaticFiles(w, r)

		return
	}

	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	log.Debug(path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))

		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}
