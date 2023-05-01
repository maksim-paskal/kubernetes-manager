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
	"strconv"
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
	Profile  string
	Services string
	User     string
	Cluster  string
}

func (input *StartNewEnvironmentInput) Validation() error {
	if len(input.Cluster) == 0 {
		return errors.Wrap(errCreateNewBranchMissingInput, "cluster is required")
	}

	if len(input.Services) == 0 {
		return errors.Wrap(errCreateNewBranchMissingInput, "services is required")
	}

	if len(input.User) == 0 {
		return errors.Wrap(errCreateNewBranchMissingInput, "user is required")
	}

	if len(input.Profile) == 0 {
		return errors.Wrap(errCreateNewBranchMissingInput, "profile is required")
	}

	if input.GetProfile() == nil {
		return errors.Wrapf(errCreateNewBranchMissingInput, "profile %s unknown", input.Profile)
	}

	if config.GetKubernetesEndpointByName(input.Cluster) == nil {
		return errors.Wrapf(errCreateNewBranchMissingInput, "cluster %s unknown", input.Cluster)
	}

	services, err := ParseEnvironmentServices(input.Services)
	if err != nil {
		return errors.Wrap(errCreateNewBranchMissingInput, "can not get services")
	}

	selectedProjectIDs := make([]string, 0)
	for _, service := range services {
		selectedProjectIDs = append(selectedProjectIDs, service.GeProjectID())
	}

	for _, required := range input.GetProfile().GetRequired() {
		if !utils.StringInSlice(required, selectedProjectIDs) {
			return errors.Wrapf(errCreateNewBranchMissingInput, "required service is missing")
		}
	}

	return nil
}

func (input *StartNewEnvironmentInput) GetProfile() *config.ProjectProfile {
	return config.GetProjectProfileByName(input.Profile)
}

type EnvironmentServices struct {
	ProjectID int
	Ref       string
}

func (services *EnvironmentServices) GeProjectID() string {
	return strconv.Itoa(services.ProjectID)
}

func ParseEnvironmentServices(services string) ([]*EnvironmentServices, error) {
	result := make([]*EnvironmentServices, 0)

	for _, service := range strings.Split(services, ";") {
		serviceData := strings.Split(service, ":")
		if len(serviceData) != config.KeyValueLength {
			return nil, errors.Wrap(errCreateNewBranchWrongFormat, service)
		}

		projectID, err := strconv.Atoi(serviceData[0])
		if err != nil {
			return nil, errors.Wrapf(err, "error converting project id %s", serviceData[0])
		}

		result = append(result, &EnvironmentServices{
			ProjectID: projectID,
			Ref:       serviceData[1],
		})
	}

	return result, nil
}

func StartNewEnvironment(ctx context.Context, input *StartNewEnvironmentInput) (*Environment, error) {
	if len(input.Cluster) == 0 {
		input.Cluster = config.Get().KubernetesEndpoints[0].Name
	}

	environment, err := processCreateNewBranch(ctx, input)
	if err != nil {
		return nil, errors.Wrap(err, "error creating new namespace")
	}

	return environment, nil
}

func processCreateNewBranch(ctx context.Context, input *StartNewEnvironmentInput) (*Environment, error) {
	if err := input.Validation(); err != nil {
		return nil, errors.Wrap(err, "error validating")
	}

	namespace, err := GetNamespaceByServices(input.GetProfile(), input.Services)
	if err != nil {
		return nil, errors.Wrap(err, "error getting namespace")
	}

	ID := fmt.Sprintf("%s:%s", input.Cluster, namespace)

	environment, err := NewEnvironment(ctx, ID, input.User)
	if err != nil {
		return nil, errors.Wrap(err, "error creating namespace")
	}

	if err := environment.CreateGitlabPipelinesByServices(ctx, input.Services, GitlabPipelineOperationBuild); err != nil {
		return nil, errors.Wrap(err, "error creating gitlab pipelines")
	}

	return environment, nil
}

var (
	errCreateNewBranchMissingInput = errors.New("missing input")
	errCreateNewBranchWrongFormat  = errors.New("wrong format")
)

const (
	getSlugStringNamespaceLength  = 40
	getNamespaceByServicesRandLen = 10
	getNamespaceByServicesTailLen = 3
)

// Return namespace by selected profile, if profile has required services,
// namespace will have ref of first required service in namespace name.
func GetNamespaceByServices(profile *config.ProjectProfile, services string) (string, error) {
	namespaceSuffix := utils.RandomString(getNamespaceByServicesRandLen)

	if len(profile.Required) > 0 {
		environmentServices, err := ParseEnvironmentServices(services)
		if err != nil {
			return "", errors.Wrapf(err, "error parsing services")
		}

		projectForNamespace := profile.GetRequired()[0]

		for _, service := range environmentServices {
			if projectForNamespace == service.GeProjectID() {
				namespaceSuffix = fmt.Sprintf("%s-%s", service.Ref, namespaceSuffix)

				break
			}
		}
	}

	namespace := fmt.Sprintf("%s%s", profile.NamespacePrefix, namespaceSuffix)

	return sluglify.GetSlugString(
		namespace,
		getSlugStringNamespaceLength,
		utils.RandomString(getNamespaceByServicesTailLen),
	), nil
}

func NewEnvironment(ctx context.Context, id string, creator string) (*Environment, error) {
	idInfo, err := types.NewIDInfo(id)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing id")
	}

	clientset, err := client.GetClientset(idInfo.Cluster)
	if err != nil {
		return nil, errors.Wrap(err, "can not get clientset")
	}

	namespaceMeta := config.GetNamespaceMeta(idInfo.Namespace)

	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        idInfo.Namespace,
			Labels:      namespaceMeta.Labels,
			Annotations: namespaceMeta.Annotations,
		},
	}

	if namespace.ObjectMeta.Labels == nil {
		namespace.ObjectMeta.Labels = make(map[string]string)
	}

	namespace.ObjectMeta.Labels[config.Namespace] = config.TrueValue

	creatorLabel := fmt.Sprintf("%s-%s", config.LabelNamespaceCreator, creator)
	namespace.ObjectMeta.Labels[creatorLabel] = config.TrueValue

	_, err = clientset.CoreV1().Namespaces().Create(ctx, &namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error creating namespace")
	}

	environment, err := GetEnvironmentByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "error creating new namespace")
	}

	return environment, nil
}
