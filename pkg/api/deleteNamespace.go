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

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteNamespace delete kubernetes namespace.
func (e *Environment) DeleteNamespace(ctx context.Context) error {
	if e.IsSystemNamespace() {
		return errors.Wrap(errIsSystemNamespace, e.Namespace)
	}

	err := e.clientset.CoreV1().Namespaces().Delete(ctx, e.Namespace, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "error deleting namespace")
	}

	return nil
}
