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
	NamespaceDescription    string
	Links                   *config.Links
	Hosts                   []string
	HostsInternal           []string
	PodsInfo                *PodsInfo
}

// GetEnvironments list all kubernetes-manager environments.
func GetEnvironments(ctx context.Context, filter string) ([]*Environment, error) {
	result := make([]*Environment, 0)

	for cluster := range client.GetAllClientsets() {
		items, err := getEnvironmentsFromCluster(ctx, cluster, filter)
		if err != nil {
			return nil, errors.Wrap(err, "error getting namespaces in "+cluster)
		}

		result = append(result, items...)
	}

	return result, nil
}

func getEnvironmentsFromCluster(ctx context.Context, cluster string, filter string) ([]*Environment, error) {
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

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, opt)
	if err != nil {
		return nil, errors.Wrap(err, "can not get namespaces")
	}

	result := make([]*Environment, 0)

	for _, namespace := range namespaces.Items {
		item := Environment{
			clientset: clientset,
			Cluster:   cluster,
		}

		if err := item.loadFromNamespace(ctx, namespace); err != nil {
			return nil, errors.Wrap(err, "can not load namespace")
		}

		result = append(result, &item)
	}

	return result, nil
}

func (e *Environment) loadFromNamespace(ctx context.Context, namespace corev1.Namespace) error {
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
	if lastScaleDateText, ok := e.NamespaceAnnotations[config.LabelLastScaleDate]; ok {
		lastScaleDate, err := utils.StringToTime(lastScaleDateText)
		if err != nil {
			log.WithError(err).Warn("can not parse time")
		} else {
			e.NamespaceLastScaled = utils.TimeToString(lastScaleDate)
			e.NamespaceLastScaledDays = utils.DiffToNowDays(lastScaleDate)
		}
	}

	// get ingress hosts
	hosts, hostsInternal, err := e.GetHosts(ctx)
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

	// namespace description
	if description, ok := e.NamespaceAnnotations[config.LabelDescription]; ok {
		namespaceDescription, err := utils.GetTemplatedResult(description, e)
		if err != nil {
			return errors.Wrap(err, "can not parse description")
		}

		e.NamespaceDescription = string(namespaceDescription)
	}

	return nil
}

func (e *Environment) ReloadFromNamespace(ctx context.Context) error {
	namespace, err := e.clientset.CoreV1().Namespaces().Get(ctx, e.Namespace, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "can not get namespace")
	}

	return e.loadFromNamespace(ctx, *namespace)
}
