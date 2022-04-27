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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	scaleRetryMaxCount = 3
	scaleRetryTimeout  = time.Second
)

// ScaleNamespace scale deployments and statefullsets.
func ScaleNamespace(ns string, replicas int32) error {
	clientset, err := getClientset(ns)
	if err != nil {
		return errors.Wrap(err, "can not get clientset")
	}

	namespace := getNamespace(ns)

	ds, err := clientset.AppsV1().Deployments(namespace).List(Ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error listing deployments")
	}

	for _, d := range ds.Items {
		log := log.WithFields(log.Fields{
			"namespace":  ns,
			"deployment": d.Name,
		})

		dps, err := clientset.AppsV1().Deployments(namespace).Get(Ctx, d.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "error getting deployment")
		}

		dps.Spec.Replicas = &replicas

		try := 0

		for {
			_, err = clientset.AppsV1().Deployments(namespace).Update(Ctx, dps, metav1.UpdateOptions{})
			if err == nil {
				break
			}

			try++

			if try >= scaleRetryMaxCount {
				return errors.Wrap(err, "error scaling deployment")
			}

			log.WithError(err).Error()
			time.Sleep(scaleRetryTimeout)
		}
	}

	sf, err := clientset.AppsV1().StatefulSets(namespace).List(Ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error listing statefullsets")
	}

	for _, s := range sf.Items {
		log := log.WithFields(log.Fields{
			"namespace":    ns,
			"statefullset": s.Name,
		})

		ss, err := clientset.AppsV1().StatefulSets(namespace).Get(Ctx, s.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "error getting statefullset")
		}

		ss.Spec.Replicas = &replicas

		try := 0

		for {
			_, err = clientset.AppsV1().StatefulSets(namespace).Update(Ctx, ss, metav1.UpdateOptions{})
			if err == nil {
				break
			}

			try++

			if try >= scaleRetryMaxCount {
				return errors.Wrap(err, "error scaling statefullset")
			}

			log.WithError(err).Error()
			time.Sleep(scaleRetryTimeout)
		}
	}

	if replicas > 0 {
		err = SaveNamespaceAnnotation(ns, map[string]string{config.LabelLastScaleDate: utils.TimeToString(time.Now())})
		if err != nil {
			return errors.Wrap(err, "error saving lastScaleDate")
		}
	}

	return nil
}
