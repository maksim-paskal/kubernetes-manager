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
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var errCreateGitlabPipelinesByServicesError = errors.New("error creating pipelines")

// return branch name if tag exists.
func (e *Environment) createBranchIfTag(ctx context.Context, pid int, ref string) (string, error) {
	ctx, span := telemetry.Start(ctx, "api.createBranchIfTag")
	defer span.End()

	tag, tagResp, err := e.gitlabClient.Tags.GetTag(pid, ref, gitlab.WithContext(ctx))

	if tagResp.StatusCode == http.StatusNotFound {
		return ref, nil
	}

	if err != nil {
		return "", errors.Wrap(err, "can not get tag")
	}

	gitlabBranch, _, err := e.gitlabClient.Branches.CreateBranch(
		pid,
		&gitlab.CreateBranchOptions{
			Branch: gitlab.Ptr("tagfork-" + ref + "-" + utils.RandomString(config.TemporaryTokenRandLength)),
			Ref:    gitlab.Ptr(tag.Commit.ID),
		},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return "", errors.Wrap(err, "can not create branch")
	}

	return gitlabBranch.Name, nil
}

func (e *Environment) CreateGitlabPipelinesByServices(ctx context.Context, services string, op GitlabPipelineOperation) error {
	ctx, span := telemetry.Start(ctx, "api.CreateGitlabPipelinesByServices")
	defer span.End()

	if len(services) == 0 {
		return errors.New("no services was selected")
	}

	if err := op.Check(); err != nil {
		return errors.Wrap(err, "operation error")
	}

	environmentServices, err := ParseEnvironmentServices(services, nil)
	if err != nil {
		return errors.Wrap(err, "error parsing services")
	}

	// create branches for tags
	for i, environmentService := range environmentServices {
		ref, err := e.createBranchIfTag(ctx, environmentService.ProjectID, environmentService.Ref)
		if err != nil {
			return errors.Wrap(err, "error creating branch")
		}

		environmentServices[i].Ref = ref
	}

	annotations := e.NamespaceAnnotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	for _, environmentService := range environmentServices {
		label := fmt.Sprintf("%s-%d", config.LabelInstalledProject, environmentService.ProjectID)

		annotations[label] = environmentService.Ref
	}

	err = e.SaveNamespaceMeta(ctx, annotations, e.NamespaceLabels)
	if err != nil {
		return errors.Wrap(err, "error saving namespace annotations")
	}

	var (
		wg   sync.WaitGroup
		lock sync.Mutex
	)

	pipelineErrors := make([]string, 0)

	wg.Add(len(environmentServices))

	for _, environmentService := range environmentServices {
		go func(e *Environment, environmentService *EnvironmentServices) {
			defer wg.Done()

			var resultText string

			_, err := e.CreateGitlabPipeline(ctx, &CreateGitlabPipelineInput{
				ProjectID: environmentService.GeProjectID(),
				Ref:       environmentService.Ref,
				Operation: op,
			})
			if err != nil {
				resultText = err.Error()

				lock.Lock()
				defer lock.Unlock()

				pipelineErrors = append(pipelineErrors, resultText)
			}
		}(e, environmentService)
	}

	wg.Wait()

	if len(pipelineErrors) > 0 {
		return errors.Wrap(errCreateGitlabPipelinesByServicesError, strings.Join(pipelineErrors, "\n"))
	}

	return nil
}
