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

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (e *Environment) GetHosts() ([]string, error) {
	opt := metav1.ListOptions{
		LabelSelector: *config.Get().IngressFilter,
	}

	ingresss, err := e.clientset.NetworkingV1().Ingresses(e.Namespace).List(Ctx, opt)
	if err != nil {
		return nil, errors.Wrap(err, "can not get ingresses")
	}

	hosts := make([]string, 0)

	for _, ingress := range ingresss.Items {
		for _, rule := range ingress.Spec.Rules {
			host := fmt.Sprintf("%s://%s", *config.Get().IngressHostDefaultProtocol, rule.Host)
			if !utils.StringInSlice(host, hosts) {
				hosts = append(hosts, host)
			}
		}
	}

	return hosts, nil
}
