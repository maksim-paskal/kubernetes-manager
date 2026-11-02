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

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func TestIsPipelineInProgress(t *testing.T) {
	t.Parallel()

	tests := map[string]bool{
		"created":              true,
		"waiting_for_resource": true,
		"preparing":            true,
		"pending":              true,
		"running":              true,
		"scheduled":            true,
		"success":              false,
		"failed":               false,
		"canceled":             false,
		"skipped":              false,
		"":                     false,
	}

	for status, want := range tests {
		t.Run(status, func(t *testing.T) {
			t.Parallel()

			if got := isPipelineInProgress(status); got != want {
				t.Fatalf("isPipelineInProgress(%q) = %v, want %v", status, got, want)
			}
		})
	}
}

func TestSetPipelineResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status string
		want   GetGitlabPipelinesStatusResults
	}{
		{status: "running", want: GetGitlabPipelinesStatusResults{LastRunningPipeline: "url"}},
		{status: "pending", want: GetGitlabPipelinesStatusResults{LastRunningPipeline: "url"}},
		{status: "success", want: GetGitlabPipelinesStatusResults{LastSuccessPipeline: "url"}},
		{status: "failed", want: GetGitlabPipelinesStatusResults{LastErrorPipeline: "url"}},
		{status: "canceled", want: GetGitlabPipelinesStatusResults{LastErrorPipeline: "url"}},
		{status: "skipped", want: GetGitlabPipelinesStatusResults{}},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			t.Parallel()

			var got GetGitlabPipelinesStatusResults

			setPipelineResult(&got, tt.status, "url")

			if got != tt.want {
				t.Fatalf("result = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPipelineOperation(t *testing.T) {
	t.Parallel()

	variable := func(key, value string) *gitlab.PipelineVariable {
		return &gitlab.PipelineVariable{Key: key, Value: value}
	}

	tests := []struct {
		name string
		vars []*gitlab.PipelineVariable
		want GitlabPipelineOperation
	}{
		{name: "deploy", vars: []*gitlab.PipelineVariable{variable("DEPLOY", "true"), variable("NAMESPACE", "ns")}, want: GitlabPipelineOperationDeploy},
		{name: "delete", vars: []*gitlab.PipelineVariable{variable("DELETE", "true"), variable("NAMESPACE", "ns")}, want: GitlabPipelineOperationDelete},
		{name: "snapshot", vars: []*gitlab.PipelineVariable{variable("SNAPSHOT", "true")}, want: GitlabPipelineOperationSnapshot},
		{name: "operation not set to true", vars: []*gitlab.PipelineVariable{variable("DELETE", "false")}, want: ""},
		{name: "no operation", vars: []*gitlab.PipelineVariable{variable("NAMESPACE", "ns")}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := pipelineOperation(tt.vars); got != tt.want {
				t.Fatalf("operation = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsNamespacePipeline(t *testing.T) {
	t.Parallel()

	vars := []*gitlab.PipelineVariable{{Key: "NAMESPACE", Value: "ns"}}

	if !isNamespacePipeline(vars, "ns") {
		t.Fatal("pipeline of the namespace is not matched")
	}

	if isNamespacePipeline(vars, "other") {
		t.Fatal("pipeline of another namespace is matched")
	}
}
