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

	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Environment struct {
	clientset               *kubernetes.Clientset
	gitlabClient            *gitlab.Client
	ID                      string
	Cluster                 string
	Namespace               string
	NamespaceStatus         string
	NamespaceCreated        string
	NamespaceCreatedDays    int
	NamespaceLastScaled     string
	NamespaceLastScaledDays int
	NamespaceAnnotations    map[string]string
	NamespaceLabels         map[string]string
	Links                   *config.Links
	Hosts                   []string
	HostsInternal           []string
	PodsInfo                *PodsInfo
}

// GetEnvironments list all kubernetes-manager environments.
func GetEnvironments(filter string) ([]*Environment, error) {
	result := make([]*Environment, 0)

	for cluster := range client.GetAllClientsets() {
		items, err := getEnvironmentsFromCluster(cluster, filter)
		if err != nil {
			return nil, errors.Wrap(err, "error getting namespaces in "+cluster)
		}

		result = append(result, items...)
	}

	return result, nil
}

func getEnvironmentsFromCluster(cluster string, filter string) ([]*Environment, error) {
	clientset, err := client.GetClientset(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "can not get clientset")
	}

	opt := metav1.ListOptions{
		FieldSelector: "status.phase=Active",
		LabelSelector: config.FilterLabels,
	}

	if len(filter) > 0 {
		opt.LabelSelector = opt.LabelSelector + "," + filter
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(Ctx, opt)
	if err != nil {
		return nil, errors.Wrap(err, "can not get namespaces")
	}

	result := make([]*Environment, 0)

	for _, namespace := range namespaces.Items {
		item := Environment{
			clientset: clientset,
			Cluster:   cluster,
		}

		if err := item.loadFromNamespace(namespace); err != nil {
			return nil, errors.Wrap(err, "can not load namespace")
		}

		result = append(result, &item)
	}

	return result, nil
}

func (e *Environment) loadFromNamespace(namespace corev1.Namespace) error {
	if namespace.Labels == nil || namespace.Labels[config.Namespace] != config.TrueValue {
		return errors.New("namespace is not managed by kubernetes-manager")
	}

	if namespace.Status.Phase != corev1.NamespaceActive {
		return errors.New("can not get namespace, not active")
	}

	e.ID = fmt.Sprintf("%s:%s", e.Cluster, namespace.Name)
	e.gitlabClient = client.GetGitlabClient()

	e.Namespace = namespace.Name
	e.NamespaceStatus = string(namespace.Status.Phase)
	e.NamespaceCreated = utils.TimeToString(namespace.CreationTimestamp.Time)
	e.NamespaceCreatedDays = utils.DiffToNowDays(namespace.CreationTimestamp.Time)
	e.NamespaceAnnotations = namespace.Annotations
	e.NamespaceLabels = namespace.Labels

	// by default scale up is date of creation
	e.NamespaceLastScaled = e.NamespaceCreated
	e.NamespaceLastScaledDays = e.NamespaceCreatedDays

	// if namespace was manually scaleup, it use date of last scale
	if e.NamespaceAnnotations != nil && len(e.NamespaceAnnotations[config.LabelLastScaleDate]) > 0 {
		lastScaleDate, err := utils.StringToTime(namespace.GetAnnotations()[config.LabelLastScaleDate])
		if err != nil {
			log.WithError(err).Warn("can not parse time")
		} else {
			e.NamespaceLastScaled = utils.TimeToString(lastScaleDate)
			e.NamespaceLastScaledDays = utils.DiffToNowDays(lastScaleDate)
		}
	}

	// get ingress hosts
	hosts, hostsInternal, err := e.GetHosts()
	if err != nil {
		return errors.Wrap(err, "can not get environment hosts")
	}

	e.Hosts = hosts
	e.HostsInternal = hostsInternal

	// load front config
	for _, frontConfig := range GetFrontConfig().Clusters {
		if frontConfig.ClusterName == e.Cluster {
			e.Links, err = frontConfig.Links.FormatedLinks(e.Namespace)
			if err != nil {
				log.WithError(err).Error("can not get links")
			}

			break
		}
	}

	return nil
}
