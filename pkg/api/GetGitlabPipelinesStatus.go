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
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

type GetGitlabPipelinesStatusResults struct {
	LastErrorPipeline   string
	LastRunningPipeline string
	LastSuccessPipeline string
}

func GetGitlabPipelinesStatus(projectID string, ns string) (*GetGitlabPipelinesStatusResults, error) {
	if gitlabClient == nil {
		return nil, errNoGitlabClient
	}

	namespace := getNamespace(ns)

	result := GetGitlabPipelinesStatusResults{}

	lastHour := time.Now().UTC().Add(-time.Hour)
	pipelineOrderBy := "id"
	pipelineSort := "desc"

	projectPipelines, _, err := gitlabClient.Pipelines.ListProjectPipelines(projectID, &gitlab.ListProjectPipelinesOptions{
		UpdatedAfter: &lastHour,
		OrderBy:      &pipelineOrderBy,
		Sort:         &pipelineSort,
		Username:     config.Get().GitlabTokenUser,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project pipelines")
	}

	for _, projectPipeline := range projectPipelines {
		pipelineVars, _, err := gitlabClient.Pipelines.GetPipelineVariables(projectID, projectPipeline.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get project pipeline variables")
		}

		for _, pipelineVar := range pipelineVars {
			if pipelineVar.Key == gitlabNamespaceKey && pipelineVar.Value == namespace {
				if projectPipeline.Status == "running" {
					result.LastRunningPipeline = projectPipeline.WebURL

					continue
				}

				if projectPipeline.Status == "success" {
					result.LastSuccessPipeline = projectPipeline.WebURL
				}

				if projectPipeline.Status == "failed" {
					result.LastErrorPipeline = projectPipeline.WebURL
				}
			}
		}
	}

	return &result, nil
}
