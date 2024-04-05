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
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func deleteClusterRoleAndBinding(ctx context.Context, clientset *kubernetes.Clientset, roleName string, roleBindingName string) error {
	ctx, span := telemetry.Start(ctx, "api.deleteClusterRoleAndBinding")
	defer span.End()

	if strings.HasPrefix(roleName, "system:") || strings.HasPrefix(roleBindingName, "system:") {
		log.Warnf("role %s or binding %s can not be deleted", roleName, roleBindingName)

		return nil
	}

	log.Infof("deleting rolebindingname=%s,role=%s", roleBindingName, roleName)

	err := clientset.RbacV1().ClusterRoles().Delete(ctx, roleName, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "error deleting cluster role")
	}

	err = clientset.RbacV1().ClusterRoleBindings().Delete(ctx, roleBindingName, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "error deleting cluster role binding")
	}

	return nil
}

// delete all cluster role and bindings linken to namespace.
func (e *Environment) DeleteClusterRolesAndBindings(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "api.DeleteClusterRolesAndBindings")
	defer span.End()

	if e.IsSystemNamespace() {
		return errors.Wrap(errIsSystemNamespace, e.Namespace)
	}

	roleBindings, err := e.clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error getting cluster role bindings")
	}

	for _, roleBinding := range roleBindings.Items {
		if len(roleBinding.Subjects) == 1 && roleBinding.RoleRef.Kind == "ClusterRole" {
			subject := roleBinding.Subjects[0]

			if subject.Namespace == e.Namespace {
				err = deleteClusterRoleAndBinding(ctx, e.clientset, roleBinding.RoleRef.Name, roleBinding.Name)
				if err != nil {
					return errors.Wrap(err, "error deleting cluster role and binding")
				}
			}
		}
	}

	return nil
}
