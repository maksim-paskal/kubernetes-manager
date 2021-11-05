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
	Name        string
	Description string
	WebURL      string
}

func GetGitlabProjects() ([]GetGitlabProjectsItem, error) {
	git, err := gitlab.NewClient(*config.Get().GitlabToken, gitlab.WithBaseURL(*config.Get().GitlabURL))
	if err != nil {
		return nil, errors.Wrap(err, "can not connect to Gitlab")
	}

	projects, _, err := git.Projects.ListProjects(&gitlab.ListProjectsOptions{
		Topic: config.Get().ExternalServicesTopic,
	})
	if err != nil {
		return nil, errors.Wrap(err, "can not list projects")
	}

	result := make([]GetGitlabProjectsItem, len(projects))

	for i, project := range projects {
		result[i].Name = project.NameWithNamespace
		result[i].Description = project.Description
		result[i].WebURL = project.WebURL
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}
