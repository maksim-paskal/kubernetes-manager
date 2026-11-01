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
package web

import (
	"context"
	"testing"

	"github.com/maksim-paskal/kubernetes-manager/pkg/user"
)

func TestEnvironmentsScopeFilter(t *testing.T) {
	t.Parallel()

	ctx := user.Set(context.Background(), user.User{Name: "alice"})

	tests := []struct {
		name    string
		ctx     context.Context //nolint:containedctx
		scope   string
		want    string
		wantErr bool
	}{
		{name: "no scope", ctx: ctx, scope: "", want: ""},
		{name: "liked", ctx: ctx, scope: "liked", want: "kubernetes-manager/user-liked-alice=true"},
		{name: "created", ctx: ctx, scope: "created", want: "kubernetes-manager/user-creator-alice=true"},
		{name: "unknown scope", ctx: ctx, scope: "all", wantErr: true},
		{name: "no user", ctx: context.Background(), scope: "liked", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := environmentsScopeFilter(tt.ctx, tt.scope)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: %v", err)
			}

			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
