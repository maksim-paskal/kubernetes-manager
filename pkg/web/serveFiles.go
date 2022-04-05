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
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	log "github.com/sirupsen/logrus"
)

var replacer *strings.Replacer

func GetContentReplacer() *strings.Replacer {
	return strings.NewReplacer(
		"https://__setry_id__@__setry_server__/1", config.GetEnvDefault("FRONT_SENTRY_DSN", "https://id@sentry/1"),
	)
}

func serveFiles(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(*config.Get().FrontDist, filepath.Clean(r.URL.Path))

	read, err := ioutil.ReadFile(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	if replacer == nil {
		replacer = GetContentReplacer()
	}

	newContents := replacer.Replace(string(read))

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
