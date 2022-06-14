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
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateNamespace(ns string) (string, error) {
	clientset, err := getClientset(ns)
	if err != nil {
		return "", errors.Wrap(err, "can not get clientset")
	}

	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: getNamespace(ns),
			Labels: map[string]string{
				config.Namespace: "true",
			},
		},
	}

	result, err := clientset.CoreV1().Namespaces().Create(Ctx, &namespace, metav1.CreateOptions{})
	if err != nil {
		return "", errors.Wrap(err, "error creating namespace")
	}

	return result.Name, nil
}
