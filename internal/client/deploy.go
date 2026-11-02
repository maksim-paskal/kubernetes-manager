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
	"slices"

	"github.com/pkg/errors"
)

// The operations of the deploy pipelines of the services, the same as the buttons of the external services tab.
const (
	// DeployOperationBuild builds the images from the ref and then deploys them.
	DeployOperationBuild = "BUILD"
	// DeployOperationDeploy deploys without the build.
	DeployOperationDeploy = "DEPLOY"
)

// DeployServices deploys the services of the environment from their refs. The server sets the ref of every service
// in the environment and creates a GitLab pipeline for it, it answers when the pipelines are created, not finished.
func (c *Client) DeployServices(ctx context.Context, id string, services []ServiceRef, build bool) error {
	path, err := envPath(id, "make-deploy-services")
	if err != nil {
		return err
	}

	if len(services) == 0 {
		return errors.New("no services to deploy")
	}

	if err := c.checkExternalServices(ctx, id, services); err != nil {
		return err
	}

	operation := DeployOperationDeploy
	if build {
		operation = DeployOperationBuild
	}

	_, err = c.post(ctx, path, map[string]string{
		"Services":  formatServiceRefs(services),
		"Operation": operation,
	})
	if err != nil {
		return errors.Wrap(err, "failed to deploy services")
	}

	return nil
}

// checkExternalServices rejects the projects that are not external services of the environment, the server accepts any
// project id and would start a pipeline in it. The list is the same one the external services tab shows.
func (c *Client) checkExternalServices(ctx context.Context, id string, services []ServiceRef) error {
	environment, err := c.GetEnvironment(ctx, id)
	if err != nil {
		return err
	}

	known, err := c.get[[]Service](ctx, "/api/external-services", map[string]string{"namespace": environment.Namespace})
	if err != nil {
		return errors.Wrap(err, "failed to get services")
	}

	for _, service := range services {
		if !slices.ContainsFunc(known, func(item Service) bool { return item.ProjectID == int64(service.ProjectID) }) {
			return errors.Errorf("project %d is not an external service of %s", service.ProjectID, id)
		}
	}

	return nil
}
