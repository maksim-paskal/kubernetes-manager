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
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	apierrorrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

// ScaleNamespace scale deployments and statefullsets.
func (e *Environment) ScaleNamespace(replicas int32) error {
	ds, err := e.clientset.AppsV1().Deployments(e.Namespace).List(Ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error listing deployments")
	}

	for _, d := range ds.Items {
		dps, err := e.clientset.AppsV1().Deployments(e.Namespace).Get(Ctx, d.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "error getting deployment")
		}

		dps.Spec.Replicas = &replicas

		err = wait.ExponentialBackoff(retry.DefaultBackoff, func() (bool, error) {
			_, err = e.clientset.AppsV1().Deployments(e.Namespace).Update(Ctx, dps, metav1.UpdateOptions{})
			switch {
			case err == nil:
				return true, nil
			case apierrorrs.IsConflict(err):
				return false, nil
			case err != nil:
				return false, errors.Wrapf(err, "failed to update deployment %s/%s", e.Namespace, dps.Name)
			}

			return false, nil
		})

		if err != nil {
			return errors.Wrap(err, "error updating deployment")
		}
	}

	sf, err := e.clientset.AppsV1().StatefulSets(e.Namespace).List(Ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error listing statefullsets")
	}

	for _, s := range sf.Items {
		ss, err := e.clientset.AppsV1().StatefulSets(e.Namespace).Get(Ctx, s.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "error getting statefullset")
		}

		ss.Spec.Replicas = &replicas

		err = wait.ExponentialBackoff(retry.DefaultBackoff, func() (bool, error) {
			_, err = e.clientset.AppsV1().StatefulSets(e.Namespace).Update(Ctx, ss, metav1.UpdateOptions{})
			switch {
			case err == nil:
				return true, nil
			case apierrorrs.IsConflict(err):
				return false, nil
			case err != nil:
				return false, errors.Wrapf(err, "failed to update statefullset %s/%s", e.Namespace, ss.Name)
			}

			return false, nil
		})

		if err != nil {
			return errors.Wrap(err, "error updating statefullset")
		}
	}

	if replicas > 0 {
		annotation := map[string]string{config.LabelLastScaleDate: utils.TimeToString(time.Now())}

		err = e.SaveNamespaceMeta(annotation, e.NamespaceLabels)
		if err != nil {
			return errors.Wrap(err, "error saving lastScaleDate")
		}
	}

	if replicas == 0 {
		if err := e.deletePodsNow(); err != nil {
			return errors.Wrap(err, "error deleting pods")
		}
	}

	return nil
}

// deletes pods with grace-period=0.
func (e *Environment) deletePodsNow() error {
	opt := metav1.ListOptions{
		FieldSelector: runningPodSelector,
	}

	pods, err := e.clientset.CoreV1().Pods(e.Namespace).List(Ctx, opt)
	if err != nil {
		return errors.Wrap(err, "error listing pods")
	}

	zero := int64(0)

	for _, pod := range pods.Items {
		err = e.clientset.CoreV1().Pods(e.Namespace).Delete(Ctx, pod.Name, metav1.DeleteOptions{
			GracePeriodSeconds: &zero,
		})
		if err != nil {
			log.WithError(err).Warn("error deleting pod")
		}
	}

	return nil
}
