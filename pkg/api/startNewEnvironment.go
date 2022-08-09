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
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/maksim-paskal/sluglify"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StartNewEnvironmentInput struct {
	Services string
	User     string
	Cluster  string
}

func StartNewEnvironment(input *StartNewEnvironmentInput) (*Environment, error) {
	if len(input.Cluster) == 0 {
		input.Cluster = config.Get().KubernetesEndpoints[0].Name
	}

	environment, err := processCreateNewBranch(input)
	if err != nil {
		return nil, errors.Wrap(err, "error creating new namespace")
	}

	return environment, nil
}

func processCreateNewBranch(input *StartNewEnvironmentInput) (*Environment, error) {
	if err := createNewBranchValidation(input); err != nil {
		return nil, errors.Wrap(err, "error validating")
	}

	ID := fmt.Sprintf("%s:%s", input.Cluster, getNamespaceByServices(input.Services))

	environment, err := NewEnvironment(ID, input.User)
	if err != nil {
		return nil, errors.Wrap(err, "error creating namespace")
	}

	if err := environment.CreateGitlabPipelinesByServices(input.Services, GitlabPipelineOperationBuild); err != nil {
		return nil, errors.Wrap(err, "error creating gitlab pipelines")
	}

	return environment, nil
}

var (
	errCreateNewBranchMissingInput           = errors.New("missing input")
	errCreateNewBranchWrongFormat            = errors.New("wrong format")
	errCreateNewBranchRequiredServiceMissing = errors.New("required service is missing")
)

func createNewBranchValidation(input *StartNewEnvironmentInput) error {
	if len(input.Cluster) == 0 {
		return errors.Wrap(errCreateNewBranchMissingInput, "cluster is empty")
	}

	if len(input.Services) == 0 {
		return errors.Wrap(errCreateNewBranchMissingInput, "services is empty")
	}

	for _, projectTemplate := range config.Get().ProjectTemplates {
		if projectTemplate.Required {
			requiredServiceFound := false

			for _, service := range strings.Split(input.Services, ";") {
				serviceData := strings.Split(service, ":")
				if len(serviceData) != config.KeyValueLength {
					return errors.Wrap(errCreateNewBranchWrongFormat, service)
				}

				if serviceData[0] == projectTemplate.ProjectID {
					requiredServiceFound = true
				}
			}

			if !requiredServiceFound {
				errorText := fmt.Sprintf("ProjectID=%s", projectTemplate.ProjectID)

				if len(projectTemplate.RequiredMessage) > 0 {
					errorText = projectTemplate.RequiredMessage
				}

				return errors.Wrap(errCreateNewBranchRequiredServiceMissing, errorText)
			}
		}
	}

	return nil
}

const (
	getSlugStringNamespaceLength  = 40
	getNamespaceByServicesRandLen = 10
	getNamespaceByServicesTailLen = 3
)

func getNamespaceByServices(services string) string {
	namespacePrefix := fmt.Sprintf("%s-", config.Namespace)
	namespaceSuffix := utils.RandomString(getNamespaceByServicesRandLen)

	for _, service := range strings.Split(services, ";") {
		serviceData := strings.Split(service, ":")

		projectTemplate := config.Get().GetProjectTemplateByProjectID(serviceData[0])

		if projectTemplate != nil {
			if len(projectTemplate.NamespacePrefix) > 0 {
				namespacePrefix = projectTemplate.NamespacePrefix
			}

			if projectTemplate.Sluglify {
				namespaceSuffix = fmt.Sprintf("%s-%s", serviceData[1], namespaceSuffix)
			}
		}
	}

	namespace := fmt.Sprintf("%s%s", namespacePrefix, namespaceSuffix)

	return sluglify.GetSlugString(
		namespace,
		getSlugStringNamespaceLength,
		utils.RandomString(getNamespaceByServicesTailLen),
	)
}

func NewEnvironment(id string, creator string) (*Environment, error) {
	idInfo, err := types.NewIDInfo(id)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing id")
	}

	clientset, err := client.GetClientset(idInfo.Cluster)
	if err != nil {
		return nil, errors.Wrap(err, "can not get clientset")
	}

	defaultMeta := config.Get().DeepCopy().NamespaceMeta

	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        idInfo.Namespace,
			Labels:      defaultMeta.Labels,
			Annotations: defaultMeta.Annotations,
		},
	}

	if namespace.ObjectMeta.Labels == nil {
		namespace.ObjectMeta.Labels = make(map[string]string)
	}

	namespace.ObjectMeta.Labels[config.Namespace] = config.TrueValue

	creatorLabel := fmt.Sprintf("%s-%s", config.LabelNamespaceCreator, creator)
	namespace.ObjectMeta.Labels[creatorLabel] = config.TrueValue

	_, err = clientset.CoreV1().Namespaces().Create(Ctx, &namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error creating namespace")
	}

	environment, err := GetEnvironmentByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "error creating new namespace")
	}

	return environment, nil
}
