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

// https://github.com/maksim-paskal/envoy-control-plane
const (
	envoyControlPlaneName = "envoy-control-plane"
	envoyControlPlaneArg  = "-ssl.no-validation=true"
)

func (e *Environment) DisableMTLS(ctx context.Context) error {
	if e.IsSystemNamespace() {
		return errors.Wrap(errIsSystemNamespace, e.Namespace)
	}

	controlPlane, err := e.clientset.AppsV1().Deployments(e.Namespace).Get(ctx, envoyControlPlaneName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "error getting deployments")
	}

	args := controlPlane.Spec.Template.Spec.Containers[0].Args

	needUpdate := true

	for _, arg := range args {
		if arg == envoyControlPlaneArg {
			needUpdate = false

			break
		}
	}

	// update deployment only arguments not exists
	if needUpdate {
		controlPlane.Spec.Template.Spec.Containers[0].Args = append(controlPlane.Spec.Template.Spec.Containers[0].Args, envoyControlPlaneArg) //nolint:lll

		_, err = e.clientset.AppsV1().Deployments(e.Namespace).Update(ctx, controlPlane, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrap(err, "error update deployment")
		}
	}

	return nil
}
