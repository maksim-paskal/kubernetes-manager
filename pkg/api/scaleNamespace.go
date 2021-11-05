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
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
		dps, err := clientset.AppsV1().Deployments(namespace).Get(Ctx, d.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "error getting deployment")
		}

		dps.Spec.Replicas = &replicas

		_, err = clientset.AppsV1().Deployments(namespace).Update(Ctx, dps, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrap(err, "error scaling deployment")
		}
	}

	sf, err := clientset.AppsV1().StatefulSets(namespace).List(Ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error listing statefullsets")
	}

	for _, s := range sf.Items {
		ss, err := clientset.AppsV1().StatefulSets(namespace).Get(Ctx, s.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "error getting deployment")
		}

		ss.Spec.Replicas = &replicas

		_, err = clientset.AppsV1().StatefulSets(namespace).Update(Ctx, ss, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrap(err, "error scaling statefullset")
		}
	}

	if replicas > 0 {
		err = saveNamespaceLastScaleDate(ns)
		if err != nil {
			return errors.Wrap(err, "error saving lastScaleDate")
		}
	}

	return nil
}

func saveNamespaceLastScaleDate(ns string) error {
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
			Annotations: map[string]string{config.LabelLastScaleDate: time.Now().Format(time.RFC3339)},
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
