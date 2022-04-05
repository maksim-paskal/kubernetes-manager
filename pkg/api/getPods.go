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
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetPodsItemContainers struct {
	ContainerName string
}

type GetPodsItem struct {
	PodName       string
	PodLabels     map[string]string
	PodContainers []GetPodsItemContainers
}

func GetPods(ns string) ([]*GetPodsItem, error) {
	clientset, err := getClientset(ns)
	if err != nil {
		return nil, errors.Wrap(err, "can not get clientset")
	}

	namespace := getNamespace(ns)

	pods, err := clientset.CoreV1().Pods(namespace).List(Ctx, metav1.ListOptions{
		FieldSelector: runningPodSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "can not list pods")
	}

	podsData := make([]*GetPodsItem, 0)

	for _, pod := range pods.Items {
		var podContainersData []GetPodsItemContainers

		for _, podContainer := range pod.Spec.Containers {
			podContainerData := GetPodsItemContainers{
				ContainerName: podContainer.Name,
			}

			podContainersData = append(podContainersData, podContainerData)
		}

		podData := GetPodsItem{
			PodName:       pod.Name,
			PodLabels:     pod.Labels,
			PodContainers: podContainersData,
		}

		podsData = append(podsData, &podData)
	}

	return podsData, nil
}
