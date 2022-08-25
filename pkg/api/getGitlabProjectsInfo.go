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

	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

type GetGitlabProjectsInfoItem struct {
	PodRunning *GetPodByImageResult
	Pipelines  *GetGitlabPipelinesStatusResults
}

func (e *Environment) GetGitlabProjectsInfo(ctx context.Context, projectID string) (*GetGitlabProjectsInfoItem, error) {
	if e.gitlabClient == nil {
		return nil, errNoGitlabClient
	}

	project, _, err := e.gitlabClient.Projects.GetProject( //nolint:contextcheck
		projectID,
		&gitlab.GetProjectOptions{},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, errors.Wrap(err, "can not get project")
	}

	result := GetGitlabProjectsInfoItem{}

	result.Pipelines, err = e.GetGitlabPipelinesStatus(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "can not get pipelines")
	}

	result.PodRunning, err = e.GetPodByImage(ctx, project.PathWithNamespace)
	if err != nil {
		return nil, errors.Wrap(err, "can not get pod images")
	}

	return &result, nil
}
