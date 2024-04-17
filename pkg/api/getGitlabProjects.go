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
	"slices"
	"sort"
	"strconv"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
)

type GetGitlabProjectsItem struct {
	ProjectID      int
	Name           string
	Description    string
	DefaultBranch  string
	WebURL         string
	TagsList       []string
	AdditionalInfo *GetGitlabProjectsInfoItem // custom field for front end
	Deploy         string                     // custom field for front end
	// custom field for front end
	RowVariant     string `json:"_rowVariant"` //nolint:tagliatelle
	Required       bool
	SelectedBranch string
	sortPriority   int
}

// get gitlab project by profile or namespace.
func GetGitlabProjects(ctx context.Context, profile string, namespace string) ([]*GetGitlabProjectsItem, error) {
	ctx, span := telemetry.Start(ctx, "api.GetGitlabProjects")
	defer span.End()

	projects, err := GetCachedGitlabProjectsByTopic(ctx, *config.Get().ExternalServicesTopic)
	if err != nil {
		return nil, errors.Wrap(err, "can not list projects")
	}

	var (
		exludeProjects  []string
		includeProjects []string
		projectProfile  *config.ProjectProfile
	)

	if len(profile) > 0 { //nolint:gocritic
		projectProfile = config.GetProjectProfileByName(profile)
	} else if len(namespace) > 0 {
		projectProfile = config.GetProjectProfileByNamespace(namespace)
	} else {
		return nil, errors.Wrap(err, "need profile or namespace")
	}

	if projectProfile == nil {
		return nil, errors.Errorf("unknown project profile for %s %s", profile, namespace)
	}

	if projectProfile.Exclude == "*" {
		for _, project := range projects {
			if !projectProfile.IsProjectRequired(project.ID) {
				exludeProjects = append(exludeProjects, strconv.Itoa(project.ID))
			}
		}
	} else {
		exludeProjects = projectProfile.GetExclude()
	}

	includeProjects = projectProfile.GetInclude()

	result := make([]*GetGitlabProjectsItem, 0)

	for _, project := range projects {
		item := GetGitlabProjectsItem{
			ProjectID:      project.ID,
			Name:           project.NameWithNamespace,
			Description:    project.Description,
			DefaultBranch:  project.DefaultBranch,
			WebURL:         project.WebURL,
			TagsList:       formatProjectTags(project.TagList),
			Required:       projectProfile.IsProjectRequired(project.ID),
			SelectedBranch: projectProfile.GetProjectSelectedBranch(project.ID),
			sortPriority:   projectProfile.GetProjectSortPriority(project.ID),
		}

		exclude := slices.Contains(exludeProjects, strconv.Itoa(item.ProjectID))

		// if project in includes it must be always shown
		if slices.Contains(includeProjects, strconv.Itoa(item.ProjectID)) {
			exclude = false
		}

		if !exclude {
			result = append(result, &item)
		}
	}

	// sort by sortPriority and Name
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].sortPriority != result[j].sortPriority {
			return result[i].sortPriority < result[j].sortPriority
		}

		return result[i].Name < result[j].Name
	})

	return result, nil
}

func formatProjectTags(tags []string) []string {
	formatedTags := make([]string, 0)

	for _, tag := range tags {
		// ignore filter tag
		if tag == *config.Get().ExternalServicesTopic {
			continue
		}

		formatedTags = append(formatedTags, tag)
	}

	return formatedTags
}
