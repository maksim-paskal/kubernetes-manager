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

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
)

func SetRemoteServerLabels(ctx context.Context, server *hcloud.Server, newLabels map[string]string) error {
	ctx, span := telemetry.Start(ctx, "api.SetRemoteServerLabels")
	defer span.End()

	labels := server.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	for k, v := range newLabels {
		labels[k] = v
	}

	opts := hcloud.ServerUpdateOpts{
		Labels: labels,
	}

	hcloundClient := client.GetHcloudClient()

	_, _, err := hcloundClient.Server.Update(ctx, server, opts)
	if err != nil {
		return errors.Wrap(err, "error updating server")
	}

	return nil
}
