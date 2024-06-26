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

	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func handlerUser(w http.ResponseWriter, r *http.Request) {
	_, span := telemetry.Start(r.Context(), "handlerUser")
	defer span.End()

	_, err := w.Write([]byte(`{"user":"kubernetes-manager.test.com","email":"kubernetes-manager@domain.com","groups":["kubernetes-manager-admin"]}`))
	if err != nil {
		log.WithError(err).Error("Error writing response")
	}
}
