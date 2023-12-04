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
	"fmt"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	HostTypeInternal = "internal"
)

func (e *Environment) GetHosts(ctx context.Context) ([]string, []string, error) {
	ctx, span := telemetry.Start(ctx, "api.GetHosts")
	defer span.End()

	opt := metav1.ListOptions{
		LabelSelector: config.FilterLabels,
	}

	ingresss, err := e.clientset.NetworkingV1().Ingresses(e.Namespace).List(ctx, opt)
	if err != nil {
		return nil, nil, errors.Wrap(err, "can not get ingresses")
	}

	hostsDefaults := make([]string, 0)
	hostsInternal := make([]string, 0)

	for _, ingress := range ingresss.Items {
		for _, rule := range ingress.Spec.Rules {
			host := fmt.Sprintf("%s://%s", *config.Get().IngressHostDefaultProtocol, rule.Host)
			if ingress.Annotations != nil && ingress.Annotations[config.LabelType] == HostTypeInternal {
				if !utils.StringInSlice(host, hostsInternal) {
					hostsInternal = append(hostsInternal, host)
				}
			} else {
				if !utils.StringInSlice(host, hostsDefaults) {
					hostsDefaults = append(hostsDefaults, host)
				}
			}
		}
	}

	return hostsDefaults, hostsInternal, nil
}
