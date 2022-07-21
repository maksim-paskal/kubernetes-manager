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

	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	log "github.com/sirupsen/logrus"
)

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
