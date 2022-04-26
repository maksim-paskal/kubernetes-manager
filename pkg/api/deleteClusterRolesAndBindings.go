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

	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// remove orphaned cluster role that not linked to namespace.
func RemoveOrphanedClusterRoles() error {
	ctx := context.Background()

	for _, clientset := range clientsetCluster {
		roleBindings, err := clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "error getting cluster role bindings")
		}

		for _, roleBinding := range roleBindings.Items {
			if len(roleBinding.Subjects) == 1 && roleBinding.RoleRef.Kind == "ClusterRole" {
				subject := roleBinding.Subjects[0]

				if len(subject.Namespace) > 0 {
					_, err := clientset.CoreV1().Namespaces().Get(ctx, subject.Namespace, metav1.GetOptions{})
					if err != nil && strings.Contains(err.Error(), "not found") {
						err = deleteClusterRoleAndBinding(clientset, roleBinding.RoleRef.Name, roleBinding.Name)
						if err != nil {
							return errors.Wrap(err, "error deleting cluster role and binding")
						}
					}
				}
			}
		}
	}

	return nil
}

func deleteClusterRoleAndBinding(clientset *kubernetes.Clientset, roleName string, roleBindingName string) error {
	if strings.HasPrefix(roleName, "system:") || strings.HasPrefix(roleBindingName, "system:") {
		log.Warnf("role %s or binding %s can not be deleted", roleName, roleBindingName)

		return nil
	}

	log.Infof("deleting rolebindingname=%s,role=%s", roleBindingName, roleName)

	ctx := context.Background()

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
func DeleteClusterRolesAndBindings(ns string) error {
	ctx := context.Background()

	clientset, err := getClientset(ns)
	if err != nil {
		return errors.Wrap(err, "can not get clientset")
	}

	namespace := getNamespace(ns)

	if utils.IsSystemNamespace(namespace) {
		return errors.Wrap(errIsSystemNamespace, namespace)
	}

	roleBindings, err := clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error getting cluster role bindings")
	}

	for _, roleBinding := range roleBindings.Items {
		if len(roleBinding.Subjects) == 1 && roleBinding.RoleRef.Kind == "ClusterRole" {
			subject := roleBinding.Subjects[0]

			if subject.Namespace == namespace {
				err = deleteClusterRoleAndBinding(clientset, roleBinding.RoleRef.Name, roleBinding.Name)
				if err != nil {
					return errors.Wrap(err, "error deleting cluster role and binding")
				}
			}
		}
	}

	return nil
}
