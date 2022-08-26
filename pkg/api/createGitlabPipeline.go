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
	gitlabClusterKey   = "CLUSTER"
	gitlabNamespaceKey = "NAMESPACE"
)

type GitlabPipelineOperation string

func (op GitlabPipelineOperation) Check() error {
	if len(op) == 0 {
		return errors.New("empty operation")
	}

	if op == GitlabPipelineOperationBuild {
		return nil
	}

	if op == GitlabPipelineOperationDelete {
		return nil
	}

	if op == GitlabPipelineOperationDeploy {
		return nil
	}

	return errors.Errorf("unknown operation %s", op)
}

const (
	GitlabPipelineOperationBuild    = "BUILD"
	GitlabPipelineOperationDelete   = "DELETE"
	GitlabPipelineOperationDeploy   = "DEPLOY"
	GitlabPipelineOperationSnapshot = "SNAPSHOT"
)

func (e *Environment) CreateGitlabPipeline(projectID, ref string, op GitlabPipelineOperation) (string, error) {
	if e.gitlabClient == nil {
		return "", errNoGitlabClient
	}

	projectIDInt, err := strconv.Atoi(projectID)
	if err != nil {
		return "", errors.Wrap(err, "can not convert to number")
	}

	variables := make([]*gitlab.PipelineVariableOptions, 0)

	variables = append(variables, &gitlab.PipelineVariableOptions{
		Key:          gitlab.String(string(op)),
		Value:        gitlab.String("true"),
		VariableType: gitlab.String("env_var"),
	})

	variables = append(variables, &gitlab.PipelineVariableOptions{
		Key:          gitlab.String(gitlabNamespaceKey),
		Value:        gitlab.String(e.Namespace),
		VariableType: gitlab.String("env_var"),
	})

	variables = append(variables, &gitlab.PipelineVariableOptions{
		Key:          gitlab.String(gitlabClusterKey),
		Value:        gitlab.String(e.Cluster),
		VariableType: gitlab.String("env_var"),
	})

	pipeline, _, err := e.gitlabClient.Pipelines.CreatePipeline(projectIDInt, &gitlab.CreatePipelineOptions{
		Ref:       &ref,
		Variables: &variables,
	})
	if err != nil {
		return "", errors.Wrap(err, "can not create pipeline")
	}

	return pipeline.WebURL, nil
}
