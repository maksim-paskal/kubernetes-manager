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
	"encoding/json"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// saveNamespaceAnnotation annotation anotation in namespace.
func SaveNamespaceAnnotation(ns string, annotation map[string]string) error {
	clientset, err := getClientset(ns)
	if err != nil {
		return errors.Wrap(err, "can not get clientset")
	}

	namespace := getNamespace(ns)

	type metadataStringValue struct {
		Annotations map[string]string `json:"annotations"`
	}

	type patchStringValue struct {
		Metadata metadataStringValue `json:"metadata"`
	}

	payload := patchStringValue{
		Metadata: metadataStringValue{
			Annotations: annotation,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "error marshaling payload")
	}

	namespaces := clientset.CoreV1().Namespaces()

	_, err = namespaces.Patch(Ctx, namespace, types.StrategicMergePatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}
