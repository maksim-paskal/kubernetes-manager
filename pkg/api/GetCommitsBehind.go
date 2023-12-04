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

	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

type GetCommitsBehindResult struct {
	DefaultBranch  *string
	WebURL         *string
	BranchNotFound bool
	CommitsBehind  *int
}

func GetCommitsBehind(ctx context.Context, p *gitlab.Project, projectID, branch string) (*GetCommitsBehindResult, error) { //nolint:lll
	ctx, span := telemetry.Start(ctx, "api.GetCommitsBehind")
	defer span.End()

	gitlabClient := client.GetGitlabClient()

	if gitlabClient == nil {
		return nil, errNoGitlabClient
	}

	gitlabProject := p

	if gitlabProject == nil {
		project, _, err := gitlabClient.Projects.GetProject(
			projectID,
			&gitlab.GetProjectOptions{},
			gitlab.WithContext(ctx),
		)
		if err != nil {
			return nil, errors.Wrap(err, "can not get project")
		}

		gitlabProject = project
	}

	result := GetCommitsBehindResult{
		DefaultBranch: &gitlabProject.DefaultBranch,
		WebURL:        &gitlabProject.WebURL,
	}

	if len(branch) > 0 && branch != gitlabProject.DefaultBranch {
		branchCompare, _, err := gitlabClient.Repositories.Compare(
			projectID,
			&gitlab.CompareOptions{
				From: &branch,
				To:   &gitlabProject.DefaultBranch,
			},
			gitlab.WithContext(ctx),
		)
		if err != nil {
			result.BranchNotFound = true
		} else {
			commitsBehind := len(branchCompare.Commits)
			result.CommitsBehind = &commitsBehind
		}
	}

	return &result, nil
}
