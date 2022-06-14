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
	"strings"
	"sync"

	"github.com/pkg/errors"
)

var errCreateGitlabPipelinesByServicesError = errors.New("error creating pipelines")

func CreateGitlabPipelinesByServices(ns, services string) error {
	projectPipelineDatas := strings.Split(services, ";")

	var (
		wg   sync.WaitGroup
		lock sync.Mutex
	)

	pipelineErrors := make([]string, 0)

	wg.Add(len(projectPipelineDatas))

	for _, projectPipelineData := range projectPipelineDatas {
		data := strings.Split(projectPipelineData, ":")

		go func(ns string, projectID string, branch string) {
			defer wg.Done()

			var resultText string

			_, err := CreateGitlabPipeline(ns, projectID, branch)
			if err != nil {
				resultText = err.Error()

				lock.Lock()
				defer lock.Unlock()

				pipelineErrors = append(pipelineErrors, resultText)
			}
		}(ns, data[0], data[1])
	}

	wg.Wait()

	if len(pipelineErrors) > 0 {
		return errors.Wrap(errCreateGitlabPipelinesByServicesError, strings.Join(pipelineErrors, "\n"))
	}

	return nil
}
