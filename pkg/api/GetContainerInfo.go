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
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetContainerInfoResult struct {
	PodAnnotations map[string]string
	PodLabels      map[string]string
	ContainerImage string
}

func (e *Environment) GetContainerInfo(container string) (*GetContainerInfoResult, error) {
	containerInfo, err := types.NewContainerInfo(container)
	if err != nil {
		return nil, err
	}

	podInfo, err := e.clientset.CoreV1().Pods(e.Namespace).Get(Ctx, containerInfo.PodName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var podContainer corev1.Container

	for _, c := range podInfo.Spec.Containers {
		if c.Name == containerInfo.ContainerName {
			podContainer = c

			break
		}
	}

	if podContainer.Name != containerInfo.ContainerName {
		return nil, errors.New("container not found")
	}

	containerInfoResults := GetContainerInfoResult{
		PodAnnotations: podInfo.Annotations,
		PodLabels:      podInfo.Labels,
		ContainerImage: podContainer.Image,
	}

	return &containerInfoResults, nil
}
