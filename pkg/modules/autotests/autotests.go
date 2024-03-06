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
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
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
		CustomAction     *config.AutotestCustomAction
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
		PipelineDuration     string
		CommitShortSHA       string
		ResultURL            string
		Status               PipelineStatus
		Test                 string
		TestNamespace        string
		PipelineEnv          map[string]string
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

func (d *Details) Normalize(a *config.Autotest) error {
	if d.CustomAction.ProjectID == 0 {
		d.CustomAction.ProjectID = a.ProjectID
	}

	return nil
}

func GetAutotestDetails(ctx context.Context, environment *api.Environment, size int) (*Details, error) {
	ctx, span := telemetry.Start(ctx, "autotests.GetAutotestDetails")
	defer span.End()

	autotestConfig := config.Get().GetAutotestByID(environment.ID)

	if autotestConfig == nil {
		return nil, errNotFound
	}

	result := Details{
		CustomAction:  autotestConfig.CustomAction.DeepCopy(),
		Actions:       autotestConfig.Actions,
		Pipelines:     []*Pipeline{},
		LastPipelines: []*Pipeline{},
	}

	// add defaults values if not set
	if err := result.Normalize(autotestConfig); err != nil {
		return nil, errors.Wrap(err, "error normalizing")
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

		pipelineInfo, _, err := gitlabClient.Pipelines.GetPipeline(pipeline.ProjectID, pipeline.ID, gitlab.WithContext(ctx))
		if err != nil {
			return nil, errors.Wrap(err, "error getting pipeline info")
		}

		pipelineDuration := time.Duration(pipelineInfo.Duration) * time.Second

		item := &Pipeline{
			PipelineID:           strconv.Itoa(pipeline.ID),
			CommitShortSHA:       pipeline.SHA[:8],
			Status:               PipelineStatus(pipeline.Status),
			PipelineURL:          pipeline.WebURL,
			PipelineCreated:      utils.TimeToString(*pipeline.CreatedAt),
			PipelineCreatedHuman: utils.HumanizeDuration(utils.HumanizeDurationShort, time.Since(*pipeline.CreatedAt)),
			PipelineEnv:          make(map[string]string),
			PipelineDuration:     utils.HumanizeDuration(utils.HumanizeDurationShort, pipelineDuration),
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
			item.PipelineEnv[pipelineVariable.Key] = pipelineVariable.Value

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

		resultURL, err := utils.GetTemplatedResult(ctx, autotestConfig.ReportURL, item)
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

type StartAutotestInput struct {
	environment *api.Environment
	Ref         string
	Test        string
	Force       bool
	ExtraEnv    map[string]string
}

func (s *StartAutotestInput) GetUser(ctx context.Context) string {
	security, ok := ctx.Value(types.ContextSecurityKey).(types.ContextSecurity)
	if ok {
		return security.Owner
	}

	return ""
}

func (s *StartAutotestInput) Validate(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "autotests.Validate")
	defer span.End()

	if len(s.Test) == 0 {
		return errors.New("test type is empty")
	}

	if len(s.GetUser(ctx)) == 0 {
		return errors.New("user is empty")
	}

	if s.environment == nil {
		return errors.New("environment is empty")
	}

	return nil
}

func (s *StartAutotestInput) SetEnvironment(environment *api.Environment) {
	s.environment = environment
}

func StartAutotest(ctx context.Context, input *StartAutotestInput) error {
	ctx, span := telemetry.Start(ctx, "api.StartAutotest")
	defer span.End()

	if err := input.Validate(ctx); err != nil {
		return errors.Wrap(err, "error validating input")
	}

	autotestConfig := config.Get().GetAutotestByID(input.environment.ID)

	if autotestConfig == nil {
		return errNotFound
	}

	action := autotestConfig.GetActionByTest(input.Test)

	if action == nil {
		return errNotFound
	}

	if len(input.Ref) == 0 {
		input.Ref = action.Ref
	}

	// check for pending pipelines
	if !input.Force {
		details, err := GetAutotestDetails(ctx, input.environment, defaultGetAutotestDetailsSize)
		if err != nil {
			return errors.Wrap(err, "error getting environment details")
		}

		for _, pipeline := range details.Pipelines {
			if pipeline.Test == input.Test &&
				(pipeline.Status == pipelineStatusRunning || pipeline.Status == pipelineStatusPending) {
				return errPipelineIsRunning
			}
		}
	}

	pipelineEnv := map[string]string{
		envNameTest:      input.Test,
		envNameOwner:     input.GetUser(ctx),
		envNameNamespace: input.environment.Namespace,
	}

	if len(action.Release) > 0 {
		releaseURL, err := utils.GetTemplatedResult(ctx, action.Release, input.environment)
		if err != nil {
			return errors.Wrap(err, "error getting release url")
		}

		release, err := getReleaseName(ctx, string(releaseURL))
		if err != nil {
			return errors.Wrap(err, "error getting release name")
		}

		pipelineEnv[envNameRelease] = release
	}

	// add extra env
	for key, value := range input.ExtraEnv {
		pipelineEnv[key] = value
	}

	gitlabClient := client.GetGitlabClient()

	variables := make([]*gitlab.PipelineVariableOptions, 0)

	for key, value := range pipelineEnv {
		if len(value) == 0 {
			return errors.Errorf("env %s is empty", key)
		}

		variables = append(variables, &gitlab.PipelineVariableOptions{
			Key:          gitlab.Ptr(key),
			Value:        gitlab.Ptr(value),
			VariableType: gitlab.Ptr("env_var"),
		})
	}

	_, _, err := gitlabClient.Pipelines.CreatePipeline(
		autotestConfig.ProjectID,
		&gitlab.CreatePipelineOptions{
			Ref:       &input.Ref,
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
	ctx, span := telemetry.Start(ctx, "api.getReleaseName")
	defer span.End()

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
