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
package types

import (
	"errors"
	"strings"
)

var errIDNotCorrect = errors.New("ID not correct")

const (
	namespaceArrayItemCount = 2
	namespaceArraySplitter  = ":"
)

type Event string

const (
	EventStart Event = "start"
	EventStop  Event = "stop"
)

type WebhookMessage struct {
	Event     Event
	Cluster   string
	Namespace string
}

type IDInfo struct {
	Cluster   string
	Namespace string
}

func NewIDInfo(id string) (*IDInfo, error) {
	data := strings.Split(id, namespaceArraySplitter)

	if len(data) != namespaceArrayItemCount {
		return nil, errIDNotCorrect
	}

	return &IDInfo{Cluster: data[0], Namespace: data[1]}, nil
}
