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

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/maksim-paskal/sluglify"
)

const (
	getSlugStringNamespaceLength  = 40
	getNamespaceByServicesRandLen = 10
	getNamespaceByServicesTailLen = 3
)

func GetNamespaceByServices(services string) string {
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
