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
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

type GetGitlabProjectBranchItem struct {
	Name    string
	updated *time.Time
}

func GetGitlabProjectRefs(projectID string) ([]*GetGitlabProjectBranchItem, error) {
	if gitlabClient == nil {
		return nil, errNoGitlabClient
	}

	const (
		gitlabListPerPage = 100
		maxBranches       = 30
		maxTags           = 10
	)

	result := make([]*GetGitlabProjectBranchItem, 0)
	currentPage := 0

	// add all project branches
	for {
		currentPage++

		gitBranches, _, err := gitlabClient.Branches.ListBranches(projectID, &gitlab.ListBranchesOptions{
			ListOptions: gitlab.ListOptions{
				Page:    currentPage,
				PerPage: gitlabListPerPage,
			},
		})
		if err != nil {
			return nil, errors.Wrap(err, "can not list branches")
		}

		if len(gitBranches) == 0 {
			break
		}

		for _, gitBranch := range gitBranches {
			result = append(result, &GetGitlabProjectBranchItem{
				Name:    gitBranch.Name,
				updated: gitBranch.Commit.CommittedDate,
			})
		}
	}

	// sort branches by updated date
	sort.Slice(result, func(i, j int) bool {
		return result[i].updated.After(*result[j].updated)
	})

	// return only specified number of branches
	if len(result) > maxBranches {
		result = result[:maxBranches]
	}

	// add project tags
	orderBy := "updated"

	gitTags, _, err := gitlabClient.Tags.ListTags(projectID, &gitlab.ListTagsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    0,
			PerPage: maxTags,
		},
		OrderBy: &orderBy,
	})
	if err != nil {
		return nil, errors.Wrap(err, "can not list tags")
	}

	for _, gitTag := range gitTags {
		result = append(result, &GetGitlabProjectBranchItem{
			Name:    gitTag.Name,
			updated: gitTag.Commit.CommittedDate,
		})
	}

	return result, nil
}
