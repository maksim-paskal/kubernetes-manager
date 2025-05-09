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
	"net/http"
	"strconv"
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const (
	gitlabClusterKey     = "CLUSTER"
	gitlabNamespaceKey   = "NAMESPACE"
	gitlabVariablePrefix = "K8SMNG_"
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
	GitlabPipelineOperationBuild    GitlabPipelineOperation = "BUILD"
	GitlabPipelineOperationDelete   GitlabPipelineOperation = "DELETE"
	GitlabPipelineOperationDeploy   GitlabPipelineOperation = "DEPLOY"
	GitlabPipelineOperationSnapshot GitlabPipelineOperation = "SNAPSHOT"
)

type CreateGitlabPipelineInput struct {
	ProjectID string
	Ref       string
	Operation GitlabPipelineOperation
	Variables []*gitlab.PipelineVariableOptions
}

func (e *Environment) CreateGitlabPipeline(ctx context.Context, input *CreateGitlabPipelineInput) (string, error) {
	ctx, span := telemetry.Start(ctx, "api.CreateGitlabPipeline")
	defer span.End()

	if e.gitlabClient == nil {
		return "", errNoGitlabClient
	}

	if e.IsSystemNamespace() {
		return "", errors.New("can not create pipeline in system namespace")
	}

	projectIDInt, err := strconv.Atoi(input.ProjectID)
	if err != nil {
		return "", errors.Wrap(err, "can not convert to number")
	}

	// ensure that pipeline can be created only for branches
	_, tagResp, _ := e.gitlabClient.Tags.GetTag(projectIDInt, input.Ref, gitlab.WithContext(ctx))
	if tagResp.StatusCode != http.StatusNotFound {
		return "", errors.New("pipeline can not be created for tag")
	}

	variables := make([]*gitlab.PipelineVariableOptions, 0)

	variables = append(variables, &gitlab.PipelineVariableOptions{
		Key:          gitlab.Ptr(string(input.Operation)),
		Value:        gitlab.Ptr("true"),
		VariableType: gitlab.Ptr(gitlab.EnvVariableType),
	})

	variables = append(variables, &gitlab.PipelineVariableOptions{
		Key:          gitlab.Ptr(gitlabNamespaceKey),
		Value:        gitlab.Ptr(e.Namespace),
		VariableType: gitlab.Ptr(gitlab.EnvVariableType),
	})

	variables = append(variables, &gitlab.PipelineVariableOptions{
		Key:          gitlab.Ptr(gitlabClusterKey),
		Value:        gitlab.Ptr(e.Cluster),
		VariableType: gitlab.Ptr(gitlab.EnvVariableType),
	})

	if projectProfile := e.getProjectProfile(); projectProfile != nil {
		for key, value := range projectProfile.PipelineVariables {
			variables = append(variables, &gitlab.PipelineVariableOptions{
				Key:          gitlab.Ptr(key),
				Value:        gitlab.Ptr(value),
				VariableType: gitlab.Ptr(gitlab.EnvVariableType),
			})
		}
	}

	if clusterProfile := e.getClusterProfile(); clusterProfile != nil {
		for key, value := range clusterProfile.PipelineVariables {
			variables = append(variables, &gitlab.PipelineVariableOptions{
				Key:          gitlab.Ptr(key),
				Value:        gitlab.Ptr(value),
				VariableType: gitlab.Ptr(gitlab.EnvVariableType),
			})
		}
	}

	if len(input.Variables) > 0 {
		variables = append(variables, input.Variables...)
	}

	// add namespace annotations as pipeline variables
	for key, value := range e.NamespaceAnnotations {
		if !strings.HasPrefix(key, config.AnnotationPrefix) {
			continue
		}

		keyFormatted := strings.ReplaceAll(key, config.AnnotationPrefix, gitlabVariablePrefix)
		keyFormatted = strings.ToUpper(keyFormatted)
		keyFormatted = strings.ReplaceAll(keyFormatted, "-", "_")

		variables = append(variables, &gitlab.PipelineVariableOptions{
			Key:          gitlab.Ptr(keyFormatted),
			Value:        gitlab.Ptr(value),
			VariableType: gitlab.Ptr(gitlab.EnvVariableType),
		})
	}

	if user := e.GetUser(ctx); user != "" {
		variables = append(variables, &gitlab.PipelineVariableOptions{
			Key:          gitlab.Ptr(gitlabVariablePrefix + "USER"),
			Value:        gitlab.Ptr(user),
			VariableType: gitlab.Ptr(gitlab.EnvVariableType),
		})
	}

	pipeline, _, err := e.gitlabClient.Pipelines.CreatePipeline(
		projectIDInt,
		&gitlab.CreatePipelineOptions{
			Ref:       &input.Ref,
			Variables: &variables,
		},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return "", errors.Wrap(err, "can not create pipeline")
	}

	return pipeline.WebURL, nil
}
