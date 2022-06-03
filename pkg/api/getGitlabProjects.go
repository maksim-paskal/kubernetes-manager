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

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

type GetGitlabProjectsItem struct {
	ProjectID      int
	Name           string
	Description    string
	DefaultBranch  string
	WebURL         string
	TagsList       []string
	AdditionalInfo *GetGitlabProjectsInfoItem // custom field for front end
	Deploy         bool                       // custom field for front end
}

func GetGitlabProjects() ([]*GetGitlabProjectsItem, error) {
	if gitlabClient == nil {
		return nil, errNoGitlabClient
	}

	projects, _, err := gitlabClient.Projects.ListProjects(&gitlab.ListProjectsOptions{
		Topic: config.Get().ExternalServicesTopic,
	})
	if err != nil {
		return nil, errors.Wrap(err, "can not list projects")
	}

	result := make([]*GetGitlabProjectsItem, 0)

	for _, project := range projects {
		item := GetGitlabProjectsItem{
			ProjectID:     project.ID,
			Name:          project.NameWithNamespace,
			Description:   project.Description,
			DefaultBranch: project.DefaultBranch,
			WebURL:        project.WebURL,
			TagsList:      formatProjectTags(project.TagList),
		}

		result = append(result, &item)
	}

	sort.SliceStable(result, func(i, j int) bool {
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
