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
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

type GetGitlabPipelinesStatusResults struct {
	LastErrorPipeline   string
	LastRunningPipeline string
	LastSuccessPipeline string
}

const GetGitlabPipelinesStatusMaxLimit = 20

func (e *Environment) GetGitlabPipelinesStatus(ctx context.Context, projectID string) (*GetGitlabPipelinesStatusResults, error) { //nolint:lll
	if e.gitlabClient == nil {
		return nil, errNoGitlabClient
	}

	result := GetGitlabPipelinesStatusResults{}

	// return last 20 project pipelines, that was created by API
	projectPipelines, _, err := e.gitlabClient.Pipelines.ListProjectPipelines(
		projectID,
		&gitlab.ListProjectPipelinesOptions{
			ListOptions: gitlab.ListOptions{
				Page:    1,
				PerPage: GetGitlabPipelinesStatusMaxLimit,
			},
			Source:   gitlab.String("api"),
			OrderBy:  gitlab.String("id"),
			Sort:     gitlab.String("desc"),
			Username: config.Get().GitlabTokenUser,
		},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project pipelines")
	}

	for _, projectPipeline := range projectPipelines {
		pipelineVars, _, err := e.gitlabClient.Pipelines.GetPipelineVariables(
			projectID,
			projectPipeline.ID,
			gitlab.WithContext(ctx),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get project pipeline variables")
		}

		for _, pipelineVar := range pipelineVars {
			if pipelineVar.Key == gitlabNamespaceKey && pipelineVar.Value == e.Namespace {
				switch projectPipeline.Status {
				case "running":
					result.LastRunningPipeline = projectPipeline.WebURL
				case "success":
					result.LastSuccessPipeline = projectPipeline.WebURL
				case "failed":
					result.LastErrorPipeline = projectPipeline.WebURL
				}
				// use only first pipeline
				return &result, nil
			}
		}
	}

	return &result, nil
}
