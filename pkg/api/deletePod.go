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

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeletePod delete kubernetes pod by name or labelSelector.
func DeletePod(ns string, pod string, labelSelector string) error {
	clientset, err := getClientset(ns)
	if err != nil {
		return errors.Wrap(err, "can not get clientset")
	}

	namespace := getNamespace(ns)

	GracePeriodSeconds := int64(0)

	opt := &metav1.DeleteOptions{
		GracePeriodSeconds: &GracePeriodSeconds,
	}

	var podName string

	if len(pod) > 0 {
		podinfo := strings.Split(pod, ":")

		if len(podinfo) != config.KeyValueLength {
			return errNoPodSelected
		}

		podName = podinfo[0]
	} else {
		if len(labelSelector) < 1 {
			return errNoLabelSelector
		}

		pods, labelSelectorErr := clientset.CoreV1().Pods(namespace).List(Ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
			FieldSelector: runningPodSelector,
		})

		if labelSelectorErr != nil {
			return errors.Wrap(labelSelectorErr, "error getting pod by labels")
		}

		if len(pods.Items) == 0 {
			return errNoPodInStatusRunning
		}

		podName = pods.Items[0].Name
	}

	err = clientset.CoreV1().Pods(namespace).Delete(Ctx, podName, *opt)
	if err != nil {
		return errors.Wrap(err, "error deleting pod")
	}

	return nil
}
