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
	"github.com/pkg/errors"
)

type GetGitlabProjectsInfoItem struct {
	PodRunning     *GetPodByImageResult
	Pipelines      *GetGitlabPipelinesStatusResults
	CommitsBehind  *int
	DefaultBranch  *string
	BranchNotFound bool
}

func (e *Environment) GetGitlabProjectsInfo(ctx context.Context, projectID, branch string) (*GetGitlabProjectsInfoItem, error) { //nolint:lll
	ctx, span := telemetry.Start(ctx, "api.GetGitlabProjectsInfo")
	defer span.End()

	if e.gitlabClient == nil {
		return nil, errNoGitlabClient
	}

	project, err := GetCachedGitlabProject(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "can not get project")
	}

	result := GetGitlabProjectsInfoItem{
		DefaultBranch: &project.DefaultBranch,
	}

	result.Pipelines, err = e.GetGitlabPipelinesStatus(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "can not get pipelines")
	}

	projectImagePrefix := project.PathWithNamespace
	projectSetting := config.Get().GetProjectSetting(projectID)

	if projectSetting != nil && len(projectSetting.ImagePrefix) > 0 {
		projectImagePrefix = projectSetting.ImagePrefix
	}

	result.PodRunning, err = e.GetPodByImage(ctx, projectImagePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "can not get pod images")
	}

	getCommitsBehind, err := GetCommitsBehind(ctx, project, projectID, branch)
	if err != nil {
		return nil, errors.Wrap(err, "can not get commit behind")
	}

	result.BranchNotFound = getCommitsBehind.BranchNotFound
	result.CommitsBehind = getCommitsBehind.CommitsBehind

	return &result, nil
}
