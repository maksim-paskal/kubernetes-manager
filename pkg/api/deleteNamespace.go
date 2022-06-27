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

// DeleteNamespace delete kubernetes namespace.
func DeleteNamespace(ns string) error {
	clientset, err := getClientset(ns)
	if err != nil {
		return errors.Wrap(err, "can not get clientset")
	}

	namespace := getNamespace(ns)

	isSystemNamespace, err := IsSystemNamespace(ns)
	if err != nil {
		return errors.Wrap(err, "error getting system namespace")
	}

	if isSystemNamespace {
		return errors.Wrap(errIsSystemNamespace, namespace)
	}

	err = clientset.CoreV1().Namespaces().Delete(Ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "error deleting namespace")
	}

	return nil
}
