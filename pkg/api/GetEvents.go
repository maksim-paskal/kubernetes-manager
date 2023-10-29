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
	"sort"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetEventsResult struct {
	createdTime time.Time
	Created,
	CreatedShort,
	Type,
	Reason,
	Object,
	Message string
}

func (e *Environment) GetEvents(ctx context.Context) ([]*GetEventsResult, error) {
	events, err := e.clientset.CoreV1().Events(e.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get events")
	}

	result := make([]*GetEventsResult, len(events.Items))

	for i, event := range events.Items {
		result[i] = &GetEventsResult{
			createdTime:  event.CreationTimestamp.Time,
			Created:      utils.TimeToString(event.CreationTimestamp.Time),
			CreatedShort: utils.HumanizeDuration(utils.HumanizeDurationShort, time.Since(event.CreationTimestamp.Time)),
			Type:         event.Type,
			Reason:       event.Reason,
			Object:       fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
			Message:      event.Message,
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].createdTime.After(result[j].createdTime)
	})

	return result, nil
}
