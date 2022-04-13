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
	"strconv"

	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

const (
	gitlabBuildKey     = "BUILD"
	gitlabNamespaceKey = "NAMESPACE"
)

func CreateGitlabPipeline(ns string, projectID string, branch string) (string, error) {
	if gitlabClient == nil {
		return "", errNoGitlabClient
	}

	projectIDInt, err := strconv.Atoi(projectID)
	if err != nil {
		return "", errors.Wrap(err, "can not convert to number")
	}

	variables := make([]*gitlab.PipelineVariable, 0)

	variables = append(variables, &gitlab.PipelineVariable{
		Key:          gitlabBuildKey,
		Value:        "true",
		VariableType: "env_var",
	})

	namespace := getNamespace(ns)

	variables = append(variables, &gitlab.PipelineVariable{
		Key:          gitlabNamespaceKey,
		Value:        namespace,
		VariableType: "env_var",
	})

	pipeline, _, err := gitlabClient.Pipelines.CreatePipeline(projectIDInt, &gitlab.CreatePipelineOptions{
		Ref:       &branch,
		Variables: &variables,
	})
	if err != nil {
		return "", errors.Wrap(err, "can not create pipeline")
	}

	return pipeline.WebURL, nil
}