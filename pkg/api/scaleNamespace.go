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
	"sync"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	apierrorrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

// ScaleNamespace scale deployments and statefullsets.
func (e *Environment) ScaleNamespace(ctx context.Context, replicas int32) error {
	ctx, span := telemetry.Start(ctx, "api.ScaleNamespace")
	defer span.End()

	var wg sync.WaitGroup

	var syncErrors sync.Map

	wg.Add(3) //nolint:mnd

	go func() {
		defer wg.Done()

		if err := e.scaleDeployments(ctx, replicas); err != nil {
			syncErrors.Store("scaleDeployments", err)
			log.WithError(err).Error("error scaling deployments")
		}
	}()

	go func() {
		defer wg.Done()

		if err := e.scaleStatefulSets(ctx, replicas); err != nil {
			syncErrors.Store("scaleStatefulSets", err)
			log.WithError(err).Error("error scaling statefullsets")
		}
	}()

	go func() {
		defer wg.Done()

		if replicas > 0 {
			annotation := map[string]string{config.LabelLastScaleDate: utils.TimeToString(time.Now())}

			annotation[config.LabelScaleDownDelay] = config.Get().GetScaleDownDelay().TimeToString()

			if err := e.SaveNamespaceMeta(ctx, annotation, e.NamespaceLabels); err != nil {
				syncErrors.Store("saveNamespaceMeta", err)
				log.WithError(err).Error("error saving lastScaleDate")
			}
		}
	}()

	wg.Wait()

	syncErrorsResult := make([]error, 0)

	syncErrors.Range(func(_, value any) bool {
		if err, ok := value.(error); ok {
			syncErrorsResult = append(syncErrorsResult, err)
		}

		return true
	})

	if len(syncErrorsResult) > 0 {
		return errors.Errorf("errors: %+v", syncErrorsResult)
	}

	if replicas == 0 {
		// sometimes jobs still running after scale down
		if err := e.deleteJobs(ctx); err != nil {
			return errors.Wrap(err, "error deleting jobs")
		}

		// sometimes frezzed in Terminating state
		if err := e.deletePodsNow(ctx); err != nil {
			return errors.Wrap(err, "error deleting pods")
		}
	}

	return nil
}

// delete all jobs in namespace.
func (e *Environment) deleteJobs(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "api.deleteJobs")
	defer span.End()

	jobs, err := e.clientset.BatchV1().Jobs(e.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error listing jobs")
	}

	for _, job := range jobs.Items {
		err = e.clientset.BatchV1().Jobs(e.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{})
		if err != nil {
			log.WithError(err).Warn("error deleting job")
		}
	}

	return nil
}

// deletes all pods with grace-period=0.
func (e *Environment) deletePodsNow(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "api.deletePodsNow")
	defer span.End()

	pods, err := e.clientset.CoreV1().Pods(e.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error listing pods")
	}

	zero := int64(0)

	for _, pod := range pods.Items {
		err = e.clientset.CoreV1().Pods(e.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{
			GracePeriodSeconds: &zero,
		})
		if err != nil {
			log.WithError(err).Warn("error deleting pod")
		}
	}

	return nil
}

func (e *Environment) scaleDeployments(ctx context.Context, replicas int32) error { //nolint:dupl
	ctx, span := telemetry.Start(ctx, "api.scaleDeployments")
	defer span.End()

	ds, err := e.clientset.AppsV1().Deployments(e.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error listing deployments")
	}

	for _, d := range ds.Items {
		dps, err := e.clientset.AppsV1().Deployments(e.Namespace).Get(ctx, d.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "error getting deployment")
		}

		dps.Spec.Replicas = &replicas

		err = wait.ExponentialBackoff(retry.DefaultBackoff, func() (bool, error) {
			_, err = e.clientset.AppsV1().Deployments(e.Namespace).Update(ctx, dps, metav1.UpdateOptions{})

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

	return nil
}

func (e *Environment) scaleStatefulSets(ctx context.Context, replicas int32) error { //nolint:dupl
	ctx, span := telemetry.Start(ctx, "api.scaleStatefulSets")
	defer span.End()

	sf, err := e.clientset.AppsV1().StatefulSets(e.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error listing statefullsets")
	}

	for _, s := range sf.Items {
		ss, err := e.clientset.AppsV1().StatefulSets(e.Namespace).Get(ctx, s.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "error getting statefullset")
		}

		ss.Spec.Replicas = &replicas

		err = wait.ExponentialBackoff(retry.DefaultBackoff, func() (bool, error) {
			_, err = e.clientset.AppsV1().StatefulSets(e.Namespace).Update(ctx, ss, metav1.UpdateOptions{})

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

	return nil
}
