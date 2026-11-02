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
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
)

func TestDeployServices(t *testing.T) {
	t.Parallel()

	services := []ServiceRef{{ProjectID: 12, Ref: "fix-login"}, {ProjectID: 15, Ref: "main"}}

	tests := []struct {
		name          string
		id            string
		services      []ServiceRef
		build         bool
		status        int
		wantOperation string
		wantRequest   bool
		wantErr       bool
	}{
		{name: "build", id: "cluster:namespace", services: services, build: true, status: http.StatusOK, wantOperation: "BUILD", wantRequest: true},
		{name: "no build", id: "cluster:namespace", services: services, status: http.StatusOK, wantOperation: "DEPLOY", wantRequest: true},
		{name: "server error", id: "cluster:namespace", services: services, build: true, status: http.StatusInternalServerError, wantRequest: true, wantErr: true},
		{name: "no services", id: "cluster:namespace", build: true, wantErr: true},
		{name: "bad id", id: "namespace", services: services, build: true, wantErr: true},
		{name: "not an external service", id: "cluster:namespace", services: []ServiceRef{{ProjectID: 999, Ref: "main"}}, build: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				path string
				body map[string]string
				got  bool
			)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/cluster:namespace/info":
					_ = json.NewEncoder(w).Encode(map[string]any{"Result": types.Environment{ID: "cluster:namespace", Namespace: "namespace"}})

					return
				case "/api/external-services":
					_ = json.NewEncoder(w).Encode(map[string]any{"Result": []Service{{ProjectID: 12}, {ProjectID: 15}}})

					return
				}

				got = true
				path = r.URL.Path

				_ = json.NewDecoder(r.Body).Decode(&body)

				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			cli := NewClient().WithBaseURL(server.URL).WithBasicAuth("login", "password")

			err := cli.DeployServices(context.Background(), tt.id, tt.services, tt.build)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, want error: %v", err, tt.wantErr)
			}

			if got != tt.wantRequest {
				t.Fatalf("request sent = %v, want %v", got, tt.wantRequest)
			}

			if !tt.wantRequest || tt.wantErr {
				return
			}

			if path != "/api/cluster:namespace/make-deploy-services" {
				t.Errorf("path = %s", path)
			}

			if body["Services"] != "12:fix-login;15:main" {
				t.Errorf("services = %q", body["Services"])
			}

			if body["Operation"] != tt.wantOperation {
				t.Errorf("operation = %q, want %q", body["Operation"], tt.wantOperation)
			}
		})
	}
}
