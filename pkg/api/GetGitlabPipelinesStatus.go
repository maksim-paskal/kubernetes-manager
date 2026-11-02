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
package api

import (
	"context"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GetGitlabPipelinesStatusResults = types.GetGitlabPipelinesStatusResults

const GetGitlabPipelinesStatusMaxLimit = 20

func (e *Environment) GetGitlabPipelinesStatus(ctx context.Context, projectID string) (*GetGitlabPipelinesStatusResults, error) {
	ctx, span := telemetry.Start(ctx, "api.GetGitlabPipelinesStatus")
	defer span.End()

	if e.gitlabClient == nil {
		return nil, errNoGitlabClient
	}

	result := GetGitlabPipelinesStatusResults{}

	// return last 20 project pipelines, that was created by API
	projectPipelines, _, err := e.gitlabClient.Pipelines.ListProjectPipelines(
		projectID,
		&gitlab.ListProjectPipelinesOptions{
			Page:     1,
			PerPage:  GetGitlabPipelinesStatusMaxLimit,
			Source:   new("api"),
			OrderBy:  new("id"),
			Sort:     new("desc"),
			Username: config.Get().GitlabTokenUser,
		},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project pipelines")
	}

	for _, projectPipeline := range projectPipelines {
		pipelineVars, err := GetCachedGitlabPipelineVariables(ctx, projectID, projectPipeline.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get project pipeline variables")
		}

		if !isNamespacePipeline(pipelineVars, e.Namespace) {
			continue
		}

		operation := pipelineOperation(pipelineVars)

		// a snapshot does not deploy the services, the deploy before it is still the state
		if operation == GitlabPipelineOperationSnapshot {
			continue
		}

		// a deleted service is not in the namespace, its earlier deploys are not the state anymore
		if operation == GitlabPipelineOperationDelete && projectPipeline.Status == "success" {
			return &result, nil
		}

		setPipelineResult(&result, projectPipeline.Status, projectPipeline.WebURL)

		// use only first pipeline
		return &result, nil
	}

	return &result, nil
}

// isPipelineInProgress is true for the pipeline that is not finished yet, a queued one waits for a runner
// and is deploying as well as a running one.
func isPipelineInProgress(status string) bool {
	switch status {
	case "created", "waiting_for_resource", "preparing", "pending", "running", "scheduled":
		return true
	default:
		return false
	}
}

// setPipelineResult sets the link of the pipeline to the field of its status, a canceled pipeline is an error:
// the deploy it was meant to do did not happen.
func setPipelineResult(result *GetGitlabPipelinesStatusResults, status, webURL string) {
	switch {
	case isPipelineInProgress(status):
		result.LastRunningPipeline = webURL
	case status == "success":
		result.LastSuccessPipeline = webURL
	case status == "failed", status == "canceled":
		result.LastErrorPipeline = webURL
	}
}

// isNamespacePipeline is true for the pipeline that deploys the namespace, the namespace is a variable of it.
func isNamespacePipeline(vars []*gitlab.PipelineVariable, namespace string) bool {
	for _, v := range vars {
		if v.Key == gitlabNamespaceKey && v.Value == namespace {
			return true
		}
	}

	return false
}

// pipelineOperation returns the operation the pipeline was created with, it is the variable set to "true".
func pipelineOperation(vars []*gitlab.PipelineVariable) GitlabPipelineOperation {
	for _, op := range []GitlabPipelineOperation{
		GitlabPipelineOperationBuild,
		GitlabPipelineOperationDelete,
		GitlabPipelineOperationDeploy,
		GitlabPipelineOperationSnapshot,
	} {
		for _, v := range vars {
			if v.Key == string(op) && v.Value == "true" {
				return op
			}
		}
	}

	return ""
}
