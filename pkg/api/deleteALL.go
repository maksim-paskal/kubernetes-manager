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
	"encoding/json"
	"fmt"
)

type DeleteALLResultOperation struct {
	Result string
}

type DeleteALLResult struct {
	HasErrors                     bool
	DeleteNamespaceResult         DeleteALLResultOperation
	DeleteClusterRolesAndBindings DeleteALLResultOperation
}

func (t *DeleteALLResult) JSON() string {
	result, err := json.Marshal(t)
	if err != nil {
		return err.Error()
	}

	return string(result)
}

func (e *Environment) DeleteALL() *DeleteALLResult {
	deleteNamespace := make(chan error)
	deleteClusterRolesAndBindings := make(chan error)

	go func() {
		deleteNamespace <- e.DeleteNamespace()
	}()

	go func() {
		deleteClusterRolesAndBindings <- e.DeleteClusterRolesAndBindings()
	}()

	result := DeleteALLResult{
		DeleteNamespaceResult: DeleteALLResultOperation{
			Result: fmt.Sprintf("Namespace %s deleted", e.Namespace),
		},
		DeleteClusterRolesAndBindings: DeleteALLResultOperation{
			Result: fmt.Sprintf("Cluster role and binding in namespace %s deleted", e.Namespace),
		},
	}

	if err := <-deleteNamespace; err != nil {
		result.HasErrors = true
		result.DeleteNamespaceResult = DeleteALLResultOperation{
			Result: err.Error(),
		}
	}

	if err := <-deleteClusterRolesAndBindings; err != nil {
		result.HasErrors = true
		result.DeleteClusterRolesAndBindings = DeleteALLResultOperation{
			Result: err.Error(),
		}
	}

	return &result
}
