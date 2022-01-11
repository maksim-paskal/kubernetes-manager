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
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

// DeleteRegistryTag deletes gitlab registry tag.
func DeleteGitlabRegistryTag(tag string, projectID string) error {
	if utils.IsSystemBranch(tag) {
		return errors.Wrap(errIsSystemBranch, tag)
	}

	git, err := gitlab.NewClient(*config.Get().GitlabToken, gitlab.WithBaseURL(*config.Get().GitlabURL))
	if err != nil {
		return errors.Wrap(err, "can not connect to Gitlab API")
	}

	gitRepos, _, err := git.ContainerRegistry.ListProjectRegistryRepositories(projectID, nil)
	if err != nil {
		return errors.Wrap(err, "can list registry by projectID")
	}

	for _, gitRepo := range gitRepos {
		_, err := git.ContainerRegistry.DeleteRegistryRepositoryTag(projectID, gitRepo.ID, tag)
		if err != nil {
			return errors.Wrap(err, "error in deleting tag")
		}
	}

	return nil
}
