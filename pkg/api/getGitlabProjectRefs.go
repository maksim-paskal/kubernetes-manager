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
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

type GetGitlabProjectBranchItem struct {
	Name    string
	Slug    string
	updated *time.Time
}

type GetGitlabProjectRefsOpt struct {
	ProjectID   string
	MaxBranches int
	MaxTags     int
}

func GetGitlabProjectRefs(ctx context.Context, opts *GetGitlabProjectRefsOpt) ([]*GetGitlabProjectBranchItem, error) {
	ctx, span := telemetry.Start(ctx, "api.GetGitlabProjectRefs")
	defer span.End()

	gitlabClient := client.GetGitlabClient()

	if gitlabClient == nil {
		return nil, errNoGitlabClient
	}

	// to slug ref name - use simple logic
	// replace all unknown symbols to '-'
	slugRegexp := regexp.MustCompile(`[^a-zA-Z0-9]`)

	const (
		gitlabListPerPage = 100
	)

	result := make([]*GetGitlabProjectBranchItem, 0)
	currentPage := 0

	// add all project branches
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		currentPage++

		gitBranches, _, err := gitlabClient.Branches.ListBranches(
			opts.ProjectID,
			&gitlab.ListBranchesOptions{
				ListOptions: gitlab.ListOptions{
					Page:    currentPage,
					PerPage: gitlabListPerPage,
				},
			},
			gitlab.WithContext(ctx),
		)
		if err != nil {
			return nil, errors.Wrap(err, "can not list branches")
		}

		if len(gitBranches) == 0 {
			break
		}

		for _, gitBranch := range gitBranches {
			result = append(result, &GetGitlabProjectBranchItem{
				Name:    gitBranch.Name,
				Slug:    strings.ToLower(slugRegexp.ReplaceAllString(gitBranch.Name, "-")),
				updated: gitBranch.Commit.CommittedDate,
			})
		}
	}

	// sort branches by updated date
	sort.Slice(result, func(i, j int) bool {
		return result[i].updated.After(*result[j].updated)
	})

	// return only specified number of branches
	if len(result) > opts.MaxBranches {
		result = result[:opts.MaxBranches]
	}

	// return result if no tags specified
	if opts.MaxTags == 0 {
		return result, nil
	}

	// add project tags
	orderBy := "updated"

	gitTags, _, err := gitlabClient.Tags.ListTags(
		opts.ProjectID,
		&gitlab.ListTagsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    0,
				PerPage: opts.MaxTags,
			},
			OrderBy: &orderBy,
		},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, errors.Wrap(err, "can not list tags")
	}

	for _, gitTag := range gitTags {
		result = append(result, &GetGitlabProjectBranchItem{
			Name:    gitTag.Name,
			Slug:    strings.ToLower(slugRegexp.ReplaceAllString(gitTag.Name, "-")),
			updated: gitTag.Commit.CommittedDate,
		})
	}

	return result, nil
}
