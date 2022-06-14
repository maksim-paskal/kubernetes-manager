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
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func createNewBranch(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("createNewBranch", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	if err := checkParams(r, []string{"services"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	cluster := config.Get().KubernetesEndpoints[0].Name
	services := r.URL.Query()["services"][0]

	namespace, err := processCreateNewBranch(cluster, services)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	_, _ = w.Write([]byte(namespace))
}

var (
	errCreateNewBranchMissingParams          = errors.New("missing params")
	errCreateNewBranchWrongFormat            = errors.New("wrong format")
	errCreateNewBranchRequiredServiceMissing = errors.New("required service is missing")
)

func processCreateNewBranch(cluster, services string) (string, error) {
	if err := createNewBranchValidation(cluster, services); err != nil {
		return "", errors.Wrap(err, "error validating")
	}

	ns := fmt.Sprintf("%s:%s", cluster, api.GetNamespaceByServices(services))

	namespace, err := api.CreateNamespace(ns)
	if err != nil {
		return "", errors.Wrap(err, "error creating namespace")
	}

	if err := api.CreateGitlabPipelinesByServices(ns, services); err != nil {
		return "", errors.Wrap(err, "error creating gitlab pipelines")
	}

	return fmt.Sprintf("%s:%s", cluster, namespace), nil
}

func createNewBranchValidation(cluster, services string) error {
	if len(cluster) == 0 {
		return errors.Wrap(errCreateNewBranchMissingParams, "cluster is empty")
	}

	if len(services) == 0 {
		return errors.Wrap(errCreateNewBranchMissingParams, "services is empty")
	}

	for _, projectTemplate := range config.Get().ProjectTemplates {
		if projectTemplate.Required {
			requiredServiceFound := false

			for _, service := range strings.Split(services, ";") {
				serviceData := strings.Split(service, ":")
				if len(serviceData) != config.KeyValueLength {
					return errors.Wrap(errCreateNewBranchWrongFormat, service)
				}

				if serviceData[0] == projectTemplate.ProjectID {
					requiredServiceFound = true
				}
			}

			if !requiredServiceFound {
				errorText := fmt.Sprintf("ProjectID=%s", projectTemplate.ProjectID)

				if len(projectTemplate.RequiredMessage) > 0 {
					errorText = projectTemplate.RequiredMessage
				}

				return errors.Wrap(errCreateNewBranchRequiredServiceMissing, errorText)
			}
		}
	}

	return nil
}
