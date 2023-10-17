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
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
)

type GetFrontConfigItem struct {
	ClusterName string
	Links       *config.Links
}

type GetFrontConfigResult struct {
	Version        string
	Links          *config.Links
	Clusters       []*GetFrontConfigItem
	DebugTemplates []*config.Template
}

// Get config for front pages.
func GetFrontConfig() *GetFrontConfigResult {
	appConfig := config.Get()

	result := GetFrontConfigResult{
		Version: config.GetVersion(),
		Links:   config.Get().Links,
	}

	result.DebugTemplates = config.Get().DebugTemplates

	result.Clusters = make([]*GetFrontConfigItem, 0)

	for _, cluster := range appConfig.GetKubernetesEndpoints() {
		result.Clusters = append(result.Clusters, &GetFrontConfigItem{
			ClusterName: cluster.Name,
			Links:       cluster.Links,
		})
	}

	return &result
}
