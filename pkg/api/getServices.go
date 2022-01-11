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
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetServicesItem struct {
	Name         string
	ServiceHost  string
	ExternalName string
	Ports        string
}

// Return services and pods with port.
func GetServices(ns string) ([]GetServicesItem, error) {
	clientset, err := getClientset(ns)
	if err != nil {
		return nil, errors.Wrap(err, "can not get clientset")
	}

	namespace := getNamespace(ns)

	list, err := clientset.CoreV1().Services(namespace).List(Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error listing services")
	}

	result := make([]GetServicesItem, 0)

	for _, service := range list.Items {
		item := GetServicesItem{}

		item.Name = service.Name
		item.ServiceHost = fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace)

		if len(service.Spec.ExternalName) > 0 {
			item.ExternalName = service.Spec.ExternalName
		}

		if len(service.Spec.ExternalIPs) > 0 {
			item.ExternalName = service.Spec.ExternalIPs[0]
		}

		ports := make([]string, len(service.Spec.Ports))
		for y := range service.Spec.Ports {
			ports[y] = strconv.Itoa(int(service.Spec.Ports[y].Port))
		}

		item.Ports = strings.Join(ports, ",")

		result = append(result, item)
	}

	podList, err := clientset.CoreV1().Pods(namespace).List(Ctx, metav1.ListOptions{
		FieldSelector: runningPodSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error listing pods")
	}

	for _, pod := range podList.Items {
		ports := make([]string, 0)

		for _, container := range pod.Spec.Containers {
			for _, containerPort := range container.Ports {
				ports = append(ports, strconv.Itoa(int(containerPort.ContainerPort)))
			}
		}

		if len(ports) > 0 {
			item := GetServicesItem{
				Name:        pod.Name,
				ServiceHost: pod.Name,
				Ports:       strings.Join(ports, ","),
			}

			result = append(result, item)
		}
	}

	return result, nil
}
