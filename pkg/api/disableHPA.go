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
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DisableHPA(ns string) error {
	clientset, err := getClientset(ns)
	if err != nil {
		return errors.Wrap(err, "can not get clientset")
	}

	namespace := getNamespace(ns)

	if utils.IsSystemNamespace(namespace) {
		return errors.Wrap(errIsSystemNamespace, namespace)
	}

	hpa := clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace)

	hpas, err := hpa.List(Ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error listing hpa")
	}

	GracePeriodSeconds := int64(0)

	opt := &metav1.DeleteOptions{
		GracePeriodSeconds: &GracePeriodSeconds,
	}

	for _, hpa := range hpas.Items {
		err := clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).Delete(Ctx, hpa.Name, *opt)
		if err != nil {
			return errors.Wrap(err, "error deleting hpa")
		}
	}

	err = ScaleNamespace(ns, 1)
	if err != nil {
		return errors.Wrap(err, "error scaling namespace")
	}

	return nil
}