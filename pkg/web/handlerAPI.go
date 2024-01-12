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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/jira"
	"github.com/maksim-paskal/kubernetes-manager/pkg/metrics"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func handlerAPI(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Start(r.Context(), "handlerAPI")
	defer span.End()

	vars := mux.Vars(r)

	result, err := apiOperation(ctx, r, vars["operation"])

	if err != nil {
		span.RecordError(err)
		w.WriteHeader(http.StatusInternalServerError)

		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.WithError(err).Error()
		}

		log.
			WithError(err).
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

func apiOperation(ctx context.Context, r *http.Request, operation string) (*HandlerResult, error) { //nolint:gocyclo,maintidx,lll
	ctx, span := telemetry.Start(ctx, "web.apiOperation")
	defer span.End()

	metricsStarts := time.Now()
	defer metrics.LogRequest(operation, metricsStarts)

	result := NewHandlerResult()

	if owner := r.Header[config.HeaderOwner]; len(owner) > 0 {
		log.Infof("user %s request %s", owner[0], operation)

		ctx = context.WithValue(ctx, types.ContextSecurityKey, types.ContextSecurity{Owner: owner[0]})

		telemetry.Attributes(span, map[string]string{
			"owner": owner[0],
		})
	}

	if err := checkForMakeOperation(ctx, operation, r); err != nil {
		return result, errors.Wrap(err, "check make operation")
	}

	if err := r.ParseForm(); err != nil {
		return result, errors.Wrap(err, "can parsing request")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read request body")
	}
	defer r.Body.Close()

	switch strings.ToLower(operation) {
	case "front-config":
		result.Result = api.GetFrontConfig()
	case "environments":
		filter := r.Form.Get("filter")
		sortby := r.Form.Get("sortby")

		if len(sortby) == 0 {
			sortby = "created"
		}

		environments, err := api.GetEnvironments(ctx, filter)
		if err != nil {
			return result, err
		}

		switch sortby {
		case "created":
			// sort descending by created
			sort.Slice(environments, func(i, j int) bool {
				iCreated, _ := utils.StringToTime(environments[i].NamespaceCreated)
				jCreated, _ := utils.StringToTime(environments[j].NamespaceCreated)

				return iCreated.After(jCreated)
			})
		case "lastscaled":
			// sort descending by started
			sort.Slice(environments, func(i, j int) bool {
				iStarted, _ := utils.StringToTime(environments[i].NamespaceLastScaled)
				jStarted, _ := utils.StringToTime(environments[j].NamespaceLastScaled)

				return jStarted.After(iStarted)
			})
		default:
			return result, errors.Wrap(errBadFormat, "unknown sortby")
		}

		result.Result = environments
	case "external-services":
		profile := r.Form.Get("profile")
		namespace := r.Form.Get("namespace")

		projects, err := api.GetGitlabProjects(ctx, profile, namespace)
		if err != nil {
			return result, err
		}

		result.Result = projects
	case "project-refs":
		id := r.Form.Get("id")
		if len(id) == 0 {
			return result, errors.Wrap(errNoComandFound, "no id specified")
		}

		refs, err := api.GetGitlabProjectRefs(ctx, &api.GetGitlabProjectRefsOpt{
			ProjectID:   id,
			MaxBranches: 30, //nolint:gomnd
		})
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

		environment, err := api.StartNewEnvironment(ctx, &input)
		if err != nil {
			return result, err
		}

		result.Result = environment.ID
	case "project-profiles":
		type projectType struct {
			Name  string
			Value string
		}

		projectTypes := make([]projectType, 0)

		for _, projectProfiles := range config.Get().ProjectProfiles {
			projectTypes = append(projectTypes, projectType{
				Name:  projectProfiles.Name,
				Value: projectProfiles.Name,
			})
		}

		result.Result = projectTypes
	case "remote-servers":
		remoteServers, err := api.GetRemoteServers(ctx)
		if err != nil {
			return result, err
		}

		result.Result = remoteServers
	case "make-remote-server-action":
		input := api.SetRemoteServerActionInput{}

		err = json.Unmarshal(body, &input)
		if err != nil {
			return result, err
		}

		err := api.SetRemoteServerAction(ctx, input)
		if err != nil {
			return result, err
		}

		result.Result = "server status changed, press Refresh to view changes"

	case "make-remote-server-delay":
		input := api.SetRemoteServerDelayInput{}

		err = json.Unmarshal(body, &input)
		if err != nil {
			return result, err
		}

		err := api.SetRemoteServerDelay(ctx, input)
		if err != nil {
			return result, err
		}

		result.Result = fmt.Sprintf("Delayed scaleDown on next %s", input.Duration)
	case "jira-issue-info":
		issue := r.Form.Get("issue")
		if len(issue) == 0 {
			return result, errors.Wrap(errNoComandFound, "no issue specified")
		}

		jiraResult, err := jira.GetIssueInfo(ctx, issue)
		if err != nil {
			return result, err
		}

		result.Result = jiraResult
	case "commits-behind":
		projectID := r.Form.Get("projectID")
		branch := r.Form.Get("branch")

		if len(projectID) == 0 {
			return result, errors.Wrap(errNoComandFound, "no projectID specified")
		}

		if len(branch) == 0 {
			return result, errors.Wrap(errNoComandFound, "no branch specified")
		}

		result.Result, err = api.GetCommitsBehind(ctx, nil, projectID, branch)
		if err != nil {
			return result, err
		}

	case "cluster-info":
		clusterName := r.Form.Get("cluster")
		if len(clusterName) == 0 {
			return result, errors.Wrap(errNoComandFound, "no cluster specified")
		}

		getClusterInfo, err := api.GetClusterInfo(ctx, clusterName)
		if err != nil {
			return result, err
		}

		result.Result = getClusterInfo.ToHuman()

	default:
		return result, errors.Wrap(errNoComandFound, operation)
	}

	return result, nil
}
