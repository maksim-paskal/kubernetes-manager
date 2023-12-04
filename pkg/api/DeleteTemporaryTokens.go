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

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (e *Environment) DeleteTemporaryTokens(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "api.DeleteTemporaryTokens")
	defer span.End()

	if e.IsSystemNamespace() {
		return errors.Wrap(errIsSystemNamespace, e.Namespace)
	}

	saList, err := e.clientset.CoreV1().ServiceAccounts(e.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "kubernetes-manager=true",
	})
	if err != nil {
		return errors.Wrap(err, "error listing service account")
	}

	for _, sa := range saList.Items {
		log.Debugf("checking temporary token %s", sa.Name)

		tokenHourActive := utils.DiffToNowHours(sa.CreationTimestamp.Time)
		if tokenHourActive > config.TemporaryTokenDurationHours {
			err = e.deleteTemporaryToken(ctx, sa.Name)
			if err != nil {
				return errors.Wrap(err, "error deleting temporary token")
			}
		}
	}

	return nil
}

func (e *Environment) deleteTemporaryToken(ctx context.Context, name string) error {
	ctx, span := telemetry.Start(ctx, "api.deleteTemporaryToken")
	defer span.End()

	err := e.clientset.RbacV1().RoleBindings(e.Namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		log.WithError(err).Errorf("error deleting role binding %s", name)
	}

	err = e.clientset.RbacV1().Roles(e.Namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		log.WithError(err).Errorf("error deleting role %s", name)
	}

	err = e.clientset.CoreV1().ServiceAccounts(e.Namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "error deleting service account")
	}

	return nil
}
