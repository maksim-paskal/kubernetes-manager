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
)

type GetFrontConfigItem struct {
	ClusterName string
	Links       *config.Links
}

type GetFrontConfigBatch struct {
	ScaleDownHourMinPeriod string
	ScaleDownHourMaxPeriod string
	BatchSheduleTimezone   string
}

type GetFrontConfigResult struct {
	Version                   string
	Links                     *config.Links
	Batch                     *GetFrontConfigBatch
	Clusters                  []*GetFrontConfigItem
	ExternalServicesTemplates []*config.Template
	DebugTemplates            []*config.Template
	RemoteServersLinks        []*config.OtherLink
}

// Get config for front pages.
func GetFrontConfig() *GetFrontConfigResult {
	appConfig := config.Get()

	result := GetFrontConfigResult{
		Version: config.GetVersion(),
		Batch: &GetFrontConfigBatch{
			ScaleDownHourMinPeriod: fmt.Sprintf("%02d", config.ScaleDownHourMinPeriod),
			ScaleDownHourMaxPeriod: fmt.Sprintf("%02d", config.ScaleDownHourMaxPeriod),
			BatchSheduleTimezone:   *config.Get().BatchSheduleTimezone,
		},
		Links:              config.Get().Links,
		RemoteServersLinks: config.Get().RemoteServer.Links,
	}

	result.DebugTemplates = config.Get().DebugTemplates
	result.ExternalServicesTemplates = config.Get().ExternalServicesTemplates

	result.Clusters = make([]*GetFrontConfigItem, 0)

	for _, cluster := range appConfig.KubernetesEndpoints {
		result.Clusters = append(result.Clusters, &GetFrontConfigItem{
			ClusterName: cluster.Name,
			Links:       cluster.Links,
		})
	}

	return &result
}
