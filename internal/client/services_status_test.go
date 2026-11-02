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
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
)

func TestNewServiceStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		branch string
		info   *types.GetGitlabProjectsInfoItem
		err    error
		want   string
		wantOK bool
	}{
		{
			name:   "running pipeline wins over the old pod",
			branch: "feature",
			info: &types.GetGitlabProjectsInfoItem{
				PodRunning: &types.GetPodByImageResult{Found: true},
				Pipelines:  &types.GetGitlabPipelinesStatusResults{LastRunningPipeline: "https://gitlab/1"},
			},
			want:   ServiceDeploying,
			wantOK: true,
		},
		{
			name:   "failed pipeline wins over the old pod",
			branch: "feature",
			info: &types.GetGitlabProjectsInfoItem{
				PodRunning: &types.GetPodByImageResult{Found: true},
				Pipelines:  &types.GetGitlabPipelinesStatusResults{LastErrorPipeline: "https://gitlab/2"},
			},
			want:   ServiceFailed,
			wantOK: true,
		},
		{
			name:   "success",
			branch: "feature",
			info:   &types.GetGitlabProjectsInfoItem{Pipelines: &types.GetGitlabPipelinesStatusResults{LastSuccessPipeline: "https://gitlab/3"}},
			want:   ServiceDeployed,
			wantOK: true,
		},
		{
			name:   "first deploy has no branch yet",
			info:   &types.GetGitlabProjectsInfoItem{Pipelines: &types.GetGitlabPipelinesStatusResults{LastRunningPipeline: "https://gitlab/4"}},
			want:   ServiceDeploying,
			wantOK: true,
		},
		{
			name:   "branch without pipelines",
			branch: "main",
			info:   &types.GetGitlabProjectsInfoItem{},
			want:   ServiceDeployed,
			wantOK: true,
		},
		{
			name:   "pod without branch",
			info:   &types.GetGitlabProjectsInfoItem{PodRunning: &types.GetPodByImageResult{Found: true}},
			want:   ServiceDeployed,
			wantOK: true,
		},
		{
			name: "not in the environment",
			info: &types.GetGitlabProjectsInfoItem{PodRunning: &types.GetPodByImageResult{}},
		},
		{
			name:   "error of deployed service",
			branch: "main",
			err:    errors.New("boom"),
			want:   ServiceUnknown,
			wantOK: true,
		},
		{
			name: "error of service that is not in the environment",
			err:  errors.New("boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, ok := newServiceStatus(Service{ProjectID: 1, Description: "api"}, tt.branch, tt.info, tt.err)
			if ok != tt.wantOK || status.State != tt.want {
				t.Fatalf("status = %+v, ok = %v, want %q, %v", status, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestNewServiceStatusName(t *testing.T) {
	t.Parallel()

	status, _ := newServiceStatus(Service{Name: "group / api"}, "main", &types.GetGitlabProjectsInfoItem{}, nil)
	if status.Name != "group / api" {
		t.Fatalf("name = %q", status.Name)
	}
}

func TestNewServicesStatus(t *testing.T) {
	t.Parallel()

	svc := func(name, state string) ServiceStatus { return ServiceStatus{Name: name, State: state} }

	tests := []struct {
		name     string
		services []ServiceStatus
		want     string
	}{
		{name: "empty", want: ServicesEmpty},
		{name: "ready", services: []ServiceStatus{svc("a", ServiceDeployed), svc("b", ServiceDeployed)}, want: ServicesReady},
		{name: "deploying", services: []ServiceStatus{svc("a", ServiceDeployed), svc("b", ServiceDeploying)}, want: ServicesDeploying},
		{name: "failed wins over deploying", services: []ServiceStatus{svc("a", ServiceFailed), svc("b", ServiceDeploying)}, want: ServicesFailed},
		{name: "unknown", services: []ServiceStatus{svc("a", ServiceDeployed), svc("b", ServiceUnknown)}, want: ServicesUnknown},
		{name: "deploying wins over unknown", services: []ServiceStatus{svc("a", ServiceUnknown), svc("b", ServiceDeploying)}, want: ServicesDeploying},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := newServicesStatus("cluster:namespace", tt.services).State; got != tt.want {
				t.Fatalf("state = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetServicesStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var result any

		switch r.URL.Path {
		case "/api/cluster:namespace/info":
			result = types.Environment{
				ID:        "cluster:namespace",
				Namespace: "namespace",
				NamespaceAnnotations: map[string]string{
					"kubernetes-manager/project-1": "feature/one",
					"kubernetes-manager/project-2": "main",
				},
			}
		case "/api/external-services":
			if r.URL.Query().Get("namespace") != "namespace" {
				http.Error(w, "bad namespace", http.StatusBadRequest)

				return
			}

			result = []types.GetGitlabProjectsItem{
				{ProjectID: 1, Description: "one"},
				{ProjectID: 2, Description: "two"},
				{ProjectID: 3, Description: "three"},
			}
		case "/api/cluster:namespace/project-info":
			info := types.GetGitlabProjectsInfoItem{Pipelines: &types.GetGitlabPipelinesStatusResults{}}

			switch r.URL.Query().Get("projectID") {
			case "1":
				// the branch has a slash, it must arrive as it is
				if r.URL.Query().Get("branch") != "feature/one" {
					http.Error(w, "bad branch", http.StatusBadRequest)

					return
				}

				info.Pipelines.LastRunningPipeline = "https://gitlab/pipelines/1"
			case "2":
				info.Pipelines.LastSuccessPipeline = "https://gitlab/pipelines/2"
				info.PodRunning = &types.GetPodByImageResult{Found: true, GitHash: "abc123"}
			}

			result = info
		default:
			http.NotFound(w, r)

			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{"Result": result})
	}))
	t.Cleanup(server.Close)

	cli := NewClient().WithBaseURL(server.URL).WithBasicAuth("login", "password")

	status, err := cli.GetServicesStatus(context.Background(), "cluster:namespace")
	if err != nil {
		t.Fatal(err)
	}

	if status.State != ServicesDeploying || !slices.Equal(status.Deploying, []string{"one"}) || len(status.Failed) != 0 {
		t.Fatalf("status = %+v", status)
	}

	// the service 3 is not in the environment
	if len(status.Services) != 2 {
		t.Fatalf("services = %+v", status.Services)
	}

	if one := status.Services[0]; one.Branch != "feature/one" || one.Pipeline != "https://gitlab/pipelines/1" {
		t.Fatalf("service one = %+v", one)
	}

	if two := status.Services[1]; two.State != ServiceDeployed || two.Commit != "abc123" {
		t.Fatalf("service two = %+v", two)
	}
}
