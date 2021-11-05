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

import "encoding/json"

type DeleteALLResultOperation struct {
	Result string
}

type DeleteALLResult struct {
	DeleteNamespaceResult         DeleteALLResultOperation
	DeleteGitlabRegistryTagResult DeleteALLResultOperation
}

func (t *DeleteALLResult) JSON() string {
	result, _ := json.Marshal(t)

	return string(result)
}

func DeleteALL(ns string, tag string, projectID string) DeleteALLResult {
	deleteNamespace := make(chan error)
	deleteGitlabRegistryTag := make(chan error)

	go func() {
		deleteNamespace <- DeleteNamespace(ns)
	}()

	go func() {
		deleteGitlabRegistryTag <- DeleteGitlabRegistryTag(tag, projectID)
	}()

	result := DeleteALLResult{
		DeleteNamespaceResult: DeleteALLResultOperation{
			Result: "Namespace deleted",
		},
		DeleteGitlabRegistryTagResult: DeleteALLResultOperation{
			Result: "Registry tag deleted",
		},
	}

	if err := <-deleteNamespace; err != nil {
		result.DeleteNamespaceResult = DeleteALLResultOperation{
			Result: err.Error(),
		}
	}

	if err := <-deleteGitlabRegistryTag; err != nil {
		result.DeleteGitlabRegistryTagResult = DeleteALLResultOperation{
			Result: err.Error(),
		}
	}

	return result
}
