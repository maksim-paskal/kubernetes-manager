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
	"fmt"
	"strings"
	"sync"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/pkg/errors"
)

var errCreateGitlabPipelinesByServicesError = errors.New("error creating pipelines")

func (e *Environment) CreateGitlabPipelinesByServices(services string) error {
	if len(services) == 0 {
		return errors.New("no services was selected")
	}

	projectPipelineDatas := strings.Split(services, ";")

	annotations := e.NamespaceAnnotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	for _, projectPipelineData := range projectPipelineDatas {
		data := strings.Split(projectPipelineData, ":")

		label := fmt.Sprintf("%s-%s", config.LabelInstalledProject, data[0])

		annotations[label] = data[1]
	}

	err := e.SaveNamespaceMeta(annotations, e.NamespaceLabels)
	if err != nil {
		return errors.Wrap(err, "error saving namespace annotations")
	}

	var (
		wg   sync.WaitGroup
		lock sync.Mutex
	)

	pipelineErrors := make([]string, 0)

	wg.Add(len(projectPipelineDatas))

	for _, projectPipelineData := range projectPipelineDatas {
		data := strings.Split(projectPipelineData, ":")

		go func(e *Environment, projectID string, branch string) {
			defer wg.Done()

			var resultText string

			_, err := e.CreateGitlabPipeline(projectID, branch, GitlabPipelineOperationBuild)
			if err != nil {
				resultText = err.Error()

				lock.Lock()
				defer lock.Unlock()

				pipelineErrors = append(pipelineErrors, resultText)
			}
		}(e, data[0], data[1])
	}

	wg.Wait()

	if len(pipelineErrors) > 0 {
		return errors.Wrap(errCreateGitlabPipelinesByServicesError, strings.Join(pipelineErrors, "\n"))
	}

	return nil
}
