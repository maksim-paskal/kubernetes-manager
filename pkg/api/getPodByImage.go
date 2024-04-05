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
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
)

type GetPodByImageResult struct {
	Tag     string
	GitHash string
	Found   bool
}

func (e *Environment) GetPodByImage(ctx context.Context, imagePrefix string) (*GetPodByImageResult, error) {
	ctx, span := telemetry.Start(ctx, "api.GetPodByImage")
	defer span.End()

	pods, err := GetCachedKubernetesPodsStatus(ctx, e.Cluster, e.Namespace, PodIsRunning)
	if err != nil {
		return &GetPodByImageResult{}, errors.Wrap(err, "can not list pods")
	}

	for _, pod := range pods {
		for _, initContainer := range pod.Spec.InitContainers {
			if strings.Contains(initContainer.Image, imagePrefix) {
				return &GetPodByImageResult{
					Tag:     getImageTagName(initContainer.Image),
					GitHash: pod.Labels["githash"],
					Found:   true,
				}, nil
			}
		}

		for _, container := range pod.Spec.Containers {
			if strings.Contains(container.Image, imagePrefix) {
				return &GetPodByImageResult{
					Tag:     getImageTagName(container.Image),
					GitHash: pod.Labels["githash"],
					Found:   true,
				}, nil
			}
		}
	}

	return &GetPodByImageResult{Found: false}, nil
}

func getImageTagName(image string) string {
	imageArray := strings.Split(image, ":")

	if len(imageArray) > 0 {
		return imageArray[len(imageArray)-1]
	}

	return "latest"
}
