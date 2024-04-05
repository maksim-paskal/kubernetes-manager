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
	"encoding/json"

	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (e *Environment) SaveNamespaceMeta(ctx context.Context, annotation map[string]string, labels map[string]string) error {
	ctx, span := telemetry.Start(ctx, "api.SaveNamespaceMeta")
	defer span.End()

	type metadataStringValue struct {
		Annotations map[string]string `json:"annotations"`
		Labels      map[string]string `json:"labels"`
	}

	type patchStringValue struct {
		Metadata metadataStringValue `json:"metadata"`
	}

	payload := patchStringValue{
		Metadata: metadataStringValue{
			Annotations: annotation,
			Labels:      labels,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "error marshaling payload")
	}

	namespaces := e.clientset.CoreV1().Namespaces()

	_, err = namespaces.Patch(ctx, e.Namespace, types.StrategicMergePatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	// if OK - update labels and annotations in Environment
	for key, value := range annotation {
		e.NamespaceAnnotations[key] = value
	}

	for key, value := range labels {
		e.NamespaceLabels[key] = value
	}

	return nil
}
