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
	"testing"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
)

func TestIsLikedBy(t *testing.T) {
	t.Parallel()

	labels := map[string]string{
		config.LabelUserLiked + "-alice": config.TrueValue,
		config.LabelUserLiked + "-bob":   "false",
	}

	tests := []struct {
		name  string
		owner string
		want  bool
	}{
		{name: "liked", owner: "alice", want: true},
		{name: "unliked", owner: "bob", want: false},
		{name: "never liked", owner: "carol", want: false},
		{name: "no user", owner: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isLikedBy(labels, tt.owner); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
