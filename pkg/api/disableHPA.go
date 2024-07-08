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
	"fmt"
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (e *Environment) DisableHPA(ctx context.Context, disable bool) error {
	ctx, span := telemetry.Start(ctx, "api.DisableHPA")
	defer span.End()

	if e.IsSystemNamespace() {
		return errors.Wrap(errIsSystemNamespace, e.Namespace)
	}

	hpas, err := e.clientset.AutoscalingV1().HorizontalPodAutoscalers(e.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error listing hpa")
	}

	const disabledPrefix = "disabled-"

	for _, hpa := range hpas.Items {
		scaleTargetRefKind := strings.TrimPrefix(hpa.Spec.ScaleTargetRef.Kind, disabledPrefix)

		if disable {
			// if already disabled - skip
			if strings.HasPrefix(hpa.Spec.ScaleTargetRef.Kind, disabledPrefix) {
				continue
			}

			scaleTargetRefKind = disabledPrefix + hpa.Spec.ScaleTargetRef.Kind
		}

		payload := fmt.Sprintf(`{"spec":{"scaleTargetRef":{"kind": "%s" }}}`, scaleTargetRefKind)

		_, err := e.clientset.AutoscalingV1().HorizontalPodAutoscalers(e.Namespace).Patch(ctx,
			hpa.Name,
			types.StrategicMergePatchType,
			[]byte(payload),
			metav1.PatchOptions{},
		)
		if err != nil {
			return errors.Wrap(err, "error deleting hpa")
		}
	}

	// if we restoring original values - do nothing
	if !disable {
		return nil
	}

	err = e.ScaleNamespace(ctx, 1)
	if err != nil {
		return errors.Wrap(err, "error scaling namespace")
	}

	return nil
}
