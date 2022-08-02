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
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
)

func (e *Environment) ScaleDownDelay(durationTime time.Duration) error {
	annotation := e.NamespaceAnnotations
	if annotation == nil {
		annotation = make(map[string]string)
	}

	annotation[config.LabelScaleDownDelay] = utils.TimeToString(time.Now().Add(durationTime))

	err := e.SaveNamespaceMeta(annotation, e.NamespaceLabels)
	if err != nil {
		return err
	}

	return nil
}
