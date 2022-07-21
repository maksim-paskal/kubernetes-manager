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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func handlerAPI(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("apiHandler", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	vars := mux.Vars(r)

	result, err := apiOperation(r, vars["operation"])

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.WithError(err).Error()
		}

		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()
	} else {
		w.Header().Set("Content-Type", "application/json")

		if result.cached {
			w.Header().Set("Cache-Control", "max-age=10")
		}

		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.WithError(err).Error()
		}
	}
}

func apiOperation(r *http.Request, operation string) (*HandlerResult, error) {
	result := NewHandlerResult()

	if err := checkPOSTMethod(operation, r); err != nil {
		return result, errors.Wrap(err, "make operation must be POST")
	}

	if err := r.ParseForm(); err != nil {
		return result, errors.Wrap(err, "can parsing request")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read request body")
	}
	defer r.Body.Close()

	switch strings.ToLower(operation) {
	case "front-config":
		result.Result = api.GetFrontConfig()
	case "environments":
		filter := r.Form.Get("filter")

		environments, err := api.GetEnvironments(filter)
		if err != nil {
			return result, err
		}

		// sort descending by created
		sort.Slice(environments, func(i, j int) bool {
			iCreated, _ := utils.StringToTime(environments[i].NamespaceCreated)
			jCreated, _ := utils.StringToTime(environments[j].NamespaceCreated)

			return iCreated.After(jCreated)
		})

		result.Result = environments
	case "external-services":
		projects, err := api.GetGitlabProjects()
		if err != nil {
			return result, err
		}

		result.Result = projects
	case "project-refs":
		id := r.Form.Get("id")
		if len(id) == 0 {
			return result, errors.Wrap(errNoComandFound, "no id specified")
		}

		refs, err := api.GetGitlabProjectRefs(id)
		if err != nil {
			return result, err
		}

		refsString := make([]string, 0)

		for _, ref := range refs {
			refsString = append(refsString, ref.Name)
		}

		result.Result = refsString
	case "make-create-environment":
		input := api.StartNewEnvironmentInput{}

		err = json.Unmarshal(body, &input)
		if err != nil {
			return result, err
		}

		environment, err := api.StartNewEnvironment(&input)
		if err != nil {
			return result, err
		}

		result.Result = environment.ID
	default:
		return result, errors.Wrap(errNoComandFound, operation)
	}

	return result, nil
}
