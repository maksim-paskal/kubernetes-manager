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
package autotests

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
)

type (
	Action struct {
		Name string
		Test string
	}
	Details struct {
		Actions          []*config.AutotestAction
		Pipelines        []*Pipeline
		LastPipelines    []*Pipeline
		HasMorePipelines bool
	}
	PipelineStatus string
	Pipeline       struct {
		PipelineID           string
		PipelineURL          string
		PipelineCreated      string
		PipelineCreatedHuman string
		PipelineOwner        string
		PipelineRelease      string
		CommitShortSHA       string
		ResultURL            string
		Status               PipelineStatus
		Test                 string
		TestNamespace        string
	}
)

const (
	envNameTest      string = "TEST"
	envNameOwner     string = "OWNER"
	envNameRelease   string = "RELEASE"
	envNameNamespace string = "TEST_NAMESPACE"

	pipelineStatusSuccess PipelineStatus = "success"
	pipelineStatusRunning PipelineStatus = "running"
	pipelineStatusPending PipelineStatus = "pending"
)

var (
	errNotFound          = errors.New("for this environment autotests is not configured")
	errPipelineIsRunning = errors.New("you must stop the current pipeline before starting a new one")
)

const (
	gitlabListPerPage             = 100
	defaultGetAutotestDetailsSize = 10
)

func GetAutotestDetails(ctx context.Context, environment *api.Environment, size int) (*Details, error) {
	autotestConfig := config.Get().GetAutotestByID(environment.ID)

	if autotestConfig == nil {
		return nil, errNotFound
	}

	result := Details{
		Actions:       autotestConfig.Actions,
		Pipelines:     []*Pipeline{},
		LastPipelines: []*Pipeline{},
	}

	gitlabClient := client.GetGitlabClient()

	pipelines, _, err := gitlabClient.Pipelines.ListProjectPipelines(
		autotestConfig.ProjectID,
		&gitlab.ListProjectPipelinesOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: gitlabListPerPage,
			},
		},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error getting pipelines")
	}

	testTypes := make(map[string]bool)

	// add all valid test types for this config
	for _, action := range autotestConfig.Actions {
		testTypes[action.Test] = true
	}

	for _, pipeline := range pipelines {
		if size > 0 && len(result.Pipelines) >= size {
			// show more button in UI
			if len(result.Pipelines) < len(pipelines) {
				result.HasMorePipelines = true
			}

			break
		}

		item := &Pipeline{
			PipelineID:           strconv.Itoa(pipeline.ID),
			CommitShortSHA:       pipeline.SHA[:8],
			Status:               PipelineStatus(pipeline.Status),
			PipelineURL:          pipeline.WebURL,
			PipelineCreated:      utils.TimeToString(*pipeline.CreatedAt),
			PipelineCreatedHuman: utils.HumanizeDuration(utils.HumanizeDurationShort, time.Since(*pipeline.CreatedAt)),
		}

		pipelineVariables, _, err := gitlabClient.Pipelines.GetPipelineVariables(
			autotestConfig.ProjectID,
			pipeline.ID,
			gitlab.WithContext(ctx),
		)
		if err != nil {
			return nil, errors.Wrap(err, "error getting pipeline variables")
		}

		for _, pipelineVariable := range pipelineVariables {
			switch pipelineVariable.Key {
			case envNameTest:
				item.Test = pipelineVariable.Value
			case envNameOwner:
				item.PipelineOwner = pipelineVariable.Value
			case envNameRelease:
				item.PipelineRelease = pipelineVariable.Value
			case envNameNamespace:
				item.TestNamespace = pipelineVariable.Value
			}
		}

		// ignore pipelines with another namespace
		if autotestConfig.FilterByNamespace && item.TestNamespace != environment.Namespace {
			continue
		}

		// test type not found in config move to next pipeline
		if _, ok := testTypes[item.Test]; !ok {
			continue
		}

		resultURL, err := utils.GetTemplatedResult(autotestConfig.ReportURL, item)
		if err != nil {
			return nil, errors.Wrap(err, "error getting result url")
		}

		item.ResultURL = string(resultURL)

		result.Pipelines = append(result.Pipelines, item)
	}

	result.LastPipelines = make([]*Pipeline, 0)

	// search last pipelines for action types
	for _, action := range autotestConfig.Actions {
		for _, pipeline := range result.Pipelines {
			if pipeline.Test == action.Test && (pipeline.Status == pipelineStatusSuccess || pipeline.Status == pipelineStatusRunning) { //nolint:lll
				result.LastPipelines = append(result.LastPipelines, pipeline)

				break
			}
		}
	}

	return &result, nil
}

func StartAutotest(ctx context.Context, environment *api.Environment, test string, user string) error {
	autotestConfig := config.Get().GetAutotestByID(environment.ID)

	if autotestConfig == nil {
		return errNotFound
	}

	action := autotestConfig.GetActionByTest(test)

	if action == nil {
		return errNotFound
	}

	// check for pending pipelines
	details, err := GetAutotestDetails(ctx, environment, defaultGetAutotestDetailsSize)
	if err != nil {
		return errors.Wrap(err, "error getting environment details")
	}

	for _, pipeline := range details.Pipelines {
		if pipeline.Test == test && (pipeline.Status == pipelineStatusRunning || pipeline.Status == pipelineStatusPending) {
			return errPipelineIsRunning
		}
	}

	gitlabClient := client.GetGitlabClient()

	variables := make([]*gitlab.PipelineVariableOptions, 0)

	variables = append(variables, &gitlab.PipelineVariableOptions{
		Key:          gitlab.String(envNameTest),
		Value:        gitlab.String(test),
		VariableType: gitlab.String("env_var"),
	})

	variables = append(variables, &gitlab.PipelineVariableOptions{
		Key:          gitlab.String(envNameOwner),
		Value:        gitlab.String(user),
		VariableType: gitlab.String("env_var"),
	})

	variables = append(variables, &gitlab.PipelineVariableOptions{
		Key:          gitlab.String(envNameNamespace),
		Value:        gitlab.String(environment.Namespace),
		VariableType: gitlab.String("env_var"),
	})

	if len(action.Release) > 0 {
		releaseURL, err := utils.GetTemplatedResult(action.Release, environment)
		if err != nil {
			return errors.Wrap(err, "error getting release url")
		}

		release, err := getReleaseName(ctx, string(releaseURL))
		if err != nil {
			return errors.Wrap(err, "error getting release name")
		}

		variables = append(variables, &gitlab.PipelineVariableOptions{
			Key:          gitlab.String(envNameRelease),
			Value:        gitlab.String(release),
			VariableType: gitlab.String("env_var"),
		})
	}

	_, _, err = gitlabClient.Pipelines.CreatePipeline(
		autotestConfig.ProjectID,
		&gitlab.CreatePipelineOptions{
			Ref:       &action.Ref,
			Variables: &variables,
		},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return errors.Wrap(err, "can not create pipeline")
	}

	return nil
}

func getReleaseName(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", errors.Wrap(err, "error creating request")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "error making request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("error getting release name")
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("error getting response body")
	}

	return string(b), nil
}
