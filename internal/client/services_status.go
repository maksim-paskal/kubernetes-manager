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
package client

import (
	"context"
	"strconv"
	"sync"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/pkg/errors"
)

// Service states, the same as the icons of the external services tab: a green check, a red and a blue one.
const (
	ServiceDeployed  = "deployed"
	ServiceDeploying = "deploying"
	ServiceFailed    = "failed"
	// ServiceUnknown is a service that the server could not check.
	ServiceUnknown = "unknown"
)

// Environment services states, the first matching one wins: failed, deploying, unknown, ready.
const (
	ServicesReady     = "ready"
	ServicesDeploying = "deploying"
	ServicesFailed    = "failed"
	ServicesUnknown   = "unknown"
	// ServicesEmpty is an environment without services.
	ServicesEmpty = "empty"
)

// maxProjectInfoRequests limits the parallel requests, every one of them makes the server ask GitLab.
const maxProjectInfoRequests = 8

// projectAnnotationPrefix is the namespace annotation of the deployed branch of a project, the server sets it
// with the same constant, see config.LabelInstalledProject.
const projectAnnotationPrefix = config.LabelInstalledProject + "-"

// ServiceStatus is the state of one external service of the environment.
type ServiceStatus struct {
	ProjectID int64
	Name      string
	// Branch is the ref the service is deployed from, empty when the service was never deployed
	Branch string
	State  string
	// Pipeline is the last deploy pipeline of the service in the environment
	Pipeline string
	// Commit is the short git hash of the running pod
	Commit string
	Error  string
}

// ServicesStatus is the state of the external services of the environment.
type ServicesStatus struct {
	ID    string
	State string
	// Failed and Deploying are the names of the services in that state
	Failed    []string
	Deploying []string
	Services  []ServiceStatus
}

// GetServicesStatus returns the deploy state of the external services of the environment, the same as the tab
// external services of the web interface shows. Only the services of the environment are returned: the ones that
// have a deployed branch, a running pod or a deploy pipeline.
func (c *Client) GetServicesStatus(ctx context.Context, id string) (*ServicesStatus, error) {
	environment, err := c.GetEnvironment(ctx, id)
	if err != nil {
		return nil, err
	}

	services, err := c.get[[]Service](ctx, "/api/external-services", map[string]string{"namespace": environment.Namespace})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get services")
	}

	infoPath, err := envPath(id, "project-info")
	if err != nil {
		return nil, err
	}

	type projectInfo struct {
		info *types.GetGitlabProjectsInfoItem
		err  error
	}

	infos := make([]projectInfo, len(services))
	limit := make(chan struct{}, maxProjectInfoRequests)

	var wg sync.WaitGroup

	for i, service := range services {
		wg.Go(func() {
			limit <- struct{}{}

			defer func() { <-limit }()

			infos[i].info, infos[i].err = c.getProjectInfo(ctx, infoPath, service.ProjectID, serviceBranch(environment, service.ProjectID))
		})
	}

	wg.Wait()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := make([]ServiceStatus, 0, len(services))

	for i, service := range services {
		status, ok := newServiceStatus(service, serviceBranch(environment, service.ProjectID), infos[i].info, infos[i].err)
		if ok {
			result = append(result, status)
		}
	}

	return newServicesStatus(id, result), nil
}

func (c *Client) getProjectInfo(ctx context.Context, path string, projectID int64, branch string) (*types.GetGitlabProjectsInfoItem, error) {
	result, err := c.get[types.GetGitlabProjectsInfoItem](ctx, path, map[string]string{
		"projectID": strconv.FormatInt(projectID, 10),
		"branch":    branch,
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// serviceBranch is the deployed branch of the project, the deploy pipeline keeps it in the annotation of the namespace.
func serviceBranch(environment *Environment, projectID int64) string {
	return environment.NamespaceAnnotations[projectAnnotationPrefix+strconv.FormatInt(projectID, 10)]
}

// newServiceStatus names the state of the service from its last deploy pipeline, the pipeline is newer than the
// pod and the annotation. It returns false for a service that is not in the environment: no branch, no pod and no
// pipeline, or a failed lookup of a service without a branch.
func newServiceStatus(service Service, branch string, info *types.GetGitlabProjectsInfoItem, infoErr error) (ServiceStatus, bool) {
	status := ServiceStatus{
		ProjectID: service.ProjectID,
		Name:      service.Description,
		Branch:    branch,
	}

	if status.Name == "" {
		status.Name = service.Name
	}

	if infoErr != nil {
		// without a deployed branch nothing says the service is in the environment, most of the profile is not
		if branch == "" {
			return status, false
		}

		status.State = ServiceUnknown
		status.Error = infoErr.Error()

		return status, true
	}

	podFound := false

	if info.PodRunning != nil && info.PodRunning.Found {
		podFound = true
		status.Commit = info.PodRunning.GitHash
	}

	pipelines := info.Pipelines
	if pipelines == nil {
		pipelines = &types.GetGitlabPipelinesStatusResults{}
	}

	switch {
	case pipelines.LastRunningPipeline != "":
		status.State = ServiceDeploying
		status.Pipeline = pipelines.LastRunningPipeline
	case pipelines.LastErrorPipeline != "":
		status.State = ServiceFailed
		status.Pipeline = pipelines.LastErrorPipeline
	case pipelines.LastSuccessPipeline != "":
		status.State = ServiceDeployed
		status.Pipeline = pipelines.LastSuccessPipeline
	case branch != "" || podFound:
		status.State = ServiceDeployed
	default:
		return status, false
	}

	return status, true
}

func newServicesStatus(id string, services []ServiceStatus) *ServicesStatus {
	result := &ServicesStatus{ID: id, Failed: []string{}, Deploying: []string{}, Services: services}
	unknown := false

	for _, service := range services {
		switch service.State {
		case ServiceFailed:
			result.Failed = append(result.Failed, service.Name)
		case ServiceDeploying:
			result.Deploying = append(result.Deploying, service.Name)
		case ServiceUnknown:
			unknown = true
		}
	}

	switch {
	case len(result.Failed) > 0:
		result.State = ServicesFailed
	case len(result.Deploying) > 0:
		result.State = ServicesDeploying
	case unknown:
		result.State = ServicesUnknown
	case len(services) == 0:
		result.State = ServicesEmpty
	default:
		result.State = ServicesReady
	}

	return result
}
