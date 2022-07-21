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

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetContainersItem struct {
	Contaners []string
}

func (e *Environment) GetContainers() (*GetContainersItem, error) {
	pods, err := e.clientset.CoreV1().Pods(e.Namespace).List(Ctx, metav1.ListOptions{
		FieldSelector: runningPodSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "can not list pods")
	}

	result := GetContainersItem{
		Contaners: make([]string, 0),
	}

	for _, pod := range pods.Items {
		for _, podContainer := range pod.Spec.Containers {
			containerText := fmt.Sprintf("%s:%s", pod.Name, podContainer.Name)

			result.Contaners = append(result.Contaners, containerText)
		}
	}

	return &result, nil
}
