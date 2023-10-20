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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodContainer struct {
	Name string
}

func (e *Environment) GetPodContainers(ctx context.Context, name string) ([]*PodContainer, error) {
	pod, err := e.clientset.CoreV1().Pods(e.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := make([]*PodContainer, 0)

	podContainers := pod.Spec.Containers
	podContainers = append(podContainers, pod.Spec.InitContainers...)

	for _, container := range podContainers {
		containers = append(containers, &PodContainer{
			Name: container.Name,
		})
	}

	return containers, nil
}
