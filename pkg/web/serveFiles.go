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
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/pkg/batch"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	log "github.com/sirupsen/logrus"
)

var replacer = strings.NewReplacer(
	"__APPLICATION_VERSION__", config.GetVersion(),
	"__SCALEDOWN_MIN__", strconv.Itoa(batch.ScaleDownUtcHourMinPeriod),
	"__SCALEDOWN_MAX__", strconv.Itoa(batch.ScaleDownUtcHourMaxPeriod),
	"__FRONT_PHPMYADMIN_URL__", config.GetEnvDefault("FRONT_PHPMYADMIN_URL", ""),
	"__FRONT_DEBUG_SERVER_NAME__", config.GetEnvDefault("FRONT_DEBUG_SERVER_NAME", ""),
	"__FRONT_SENTRY_URL__", config.GetEnvDefault("FRONT_SENTRY_URL", ""),
	"__FRONT_METRICS_URL__", config.GetEnvDefault("FRONT_METRICS_URL", ""),
	"__FRONT_LOGS_URL__", config.GetEnvDefault("FRONT_LOGS_URL", ""),
	"__FRONT_TRACING_URL__", config.GetEnvDefault("FRONT_TRACING_URL", ""),
	"__FRONT_SLACK_URL__", config.GetEnvDefault("FRONT_SLACK_URL", ""),
	"__FRONT_METRICS_PATH__", config.GetEnvDefault("FRONT_METRICS_PATH", ""),
	"__FRONT_LOGS_PATH__", config.GetEnvDefault("FRONT_LOGS_PATH", ""),
	"https://__setry_id__@__setry_server__/1", config.GetEnvDefault("FRONT_SENTRY_DSN", "https://id@sentry/1"),
)

func serveFiles(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(*config.Get().FrontDist, filepath.Clean(r.URL.Path))

	mimeType := mime.TypeByExtension(filepath.Ext(path))

	if len(mimeType) > 0 {
		w.Header().Add("Content-Type", mimeType)

		if mimeType == "application/javascript" {
			w.Header().Add("Cache-Control", "max-age=31557600")
		}
	}

	read, err := ioutil.ReadFile(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	newContents := replacer.Replace(string(read))

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
