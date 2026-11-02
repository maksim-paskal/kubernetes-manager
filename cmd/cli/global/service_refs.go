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
package global

import (
	"strings"

	apiclient "github.com/maksim-paskal/kubernetes-manager/internal/client"
	"github.com/pkg/errors"
)

// ServiceRefs is the value of the repeated --service flag, each use adds one <projectID>:<ref>.
type ServiceRefs []apiclient.ServiceRef

func (s *ServiceRefs) String() string {
	items := make([]string, len(*s))

	for i, service := range *s {
		items[i] = service.String()
	}

	return strings.Join(items, ",")
}

func (s *ServiceRefs) Set(value string) error {
	service, err := apiclient.ParseServiceRef(value)
	if err != nil {
		return err
	}

	// two refs of one project would start two pipelines of it, and the last one wins in the namespace annotation
	for _, existing := range *s {
		if existing.ProjectID == service.ProjectID {
			return errors.Errorf("service %d is given twice", service.ProjectID)
		}
	}

	*s = append(*s, service)

	return nil
}

func (s *ServiceRefs) Type() string {
	return "projectID:ref"
}
