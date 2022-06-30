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
	"fmt"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetIngressList struct {
	Namespace               string
	Cluster                 string
	NamespaceName           string
	NamespaceStatus         string
	NamespaceCreated        string
	NamespaceCreatedDays    int
	NamespaceLastScaled     string
	NamespaceLastScaledDays int
	NamespaceAnotations     map[string]string
	NamespaceLabels         map[string]string
	IngressName             string
	IngressAnotations       map[string]string
	IngressLabels           map[string]string
	Hosts                   []string
	GitBranch               string
	RunningPodsCount        int
}

func (i *GetIngressList) String() string {
	out, err := json.Marshal(i)
	if err != nil {
		return err.Error()
	}

	return string(out)
}

// GetIngress list all kubernetes-manager ingresses.
func GetIngress() ([]*GetIngressList, error) {
	result := make([]*GetIngressList, 0)

	for cluster := range clientsetCluster {
		items, err := getIngressFromCluster(cluster)
		if err != nil {
			return nil, errors.Wrap(err, "error getting ingresses in "+cluster)
		}

		result = append(result, items...)
	}

	return result, nil
}

func getIngressFromCluster(cluster string) ([]*GetIngressList, error) {
	opt := metav1.ListOptions{
		LabelSelector: *config.Get().IngressFilter,
	}
	if *config.Get().IngressNoFiltration {
		opt = metav1.ListOptions{}
	}

	ingresss, err := clientsetCluster[cluster].NetworkingV1().Ingresses("").List(Ctx, opt)
	if err != nil {
		return nil, errors.Wrap(err, "can not get ingresses")
	}

	result := make([]*GetIngressList, 0)

	for _, ingress := range ingresss.Items {
		var item GetIngressList

		namespace, err := clientsetCluster[cluster].CoreV1().Namespaces().Get(Ctx, ingress.Namespace, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "can not get namespace")
		}

		item.GitBranch = ingress.Annotations[config.LabelGitBranch]

		if len(namespace.GetAnnotations()[config.LabelLastScaleDate]) > 0 {
			lastScaleDate, err := utils.StringToTime(namespace.GetAnnotations()[config.LabelLastScaleDate])
			if err != nil {
				log.WithError(err).Warn("can not parse time")
			} else {
				item.NamespaceLastScaled = utils.TimeToString(lastScaleDate)
				item.NamespaceLastScaledDays = utils.DiffToNow(lastScaleDate)
			}
		}

		item.Cluster = cluster
		item.Namespace = fmt.Sprintf("%s:%s", item.Cluster, namespace.Name)
		item.NamespaceName = namespace.Name
		item.NamespaceStatus = string(namespace.Status.Phase)
		item.NamespaceCreated = utils.TimeToString(namespace.CreationTimestamp.Time)
		item.RunningPodsCount = -1
		item.NamespaceCreatedDays = utils.DiffToNow(namespace.CreationTimestamp.Time)
		item.NamespaceAnotations = namespace.Annotations
		item.NamespaceLabels = namespace.Labels

		item.IngressName = ingress.Name
		item.IngressAnotations = ingress.Annotations
		item.IngressLabels = ingress.Labels

		for _, rule := range ingress.Spec.Rules {
			host := fmt.Sprintf("%s://%s", *config.Get().IngressHostDefaultProtocol, rule.Host)
			if !stringInSlice(host, item.Hosts) {
				item.Hosts = append(item.Hosts, host)
			}
		}

		result = append(result, &item)
	}

	return result, nil
}
