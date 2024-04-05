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
package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/cache"
	"github.com/maksim-paskal/kubernetes-manager/pkg/metrics"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
)

const (
	envJiraURL     = "JIRA_URL"
	envJiraToken   = "JIRA_TOKEN"
	requestTimeout = 30 * time.Second
)

var jiraDefaultClient = &http.Client{
	Jar:       nil,
	Timeout:   1 * time.Minute,
	Transport: metrics.NewInstrumenter("jira").InstrumentedRoundTripper(),
}

type IssueInfo struct {
	Fields struct {
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
	} `json:"fields"`
}

func GetIssueInfo(ctx context.Context, issue string) (*IssueInfo, error) {
	ctx, span := telemetry.Start(ctx, "jira.GetIssueInfo")
	defer span.End()

	cacheKey := "jira:issue:" + issue

	var cacheValue IssueInfo

	if err := cache.Client().Get(ctx, cacheKey, &cacheValue); err == nil {
		metrics.CacheHits.WithLabelValues("GetIssueInfo").Inc()

		return &cacheValue, nil
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	if _, ok := os.LookupEnv(envJiraURL); !ok {
		return nil, errors.Errorf("env %s not set", envJiraURL)
	}

	if _, ok := os.LookupEnv(envJiraToken); !ok {
		return nil, errors.Errorf("env %s not set", envJiraToken)
	}

	url := fmt.Sprintf("%s/rest/api/latest/issue/%s", os.Getenv(envJiraURL), issue)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequestWithContext")
	}

	req.Header.Set("Authorization", "Basic "+os.Getenv(envJiraToken))

	res, err := jiraDefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "jiraDefaultClient.Do")
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("invalid StatusCode res.StatusCode: %d", res.StatusCode)
	}

	defer res.Body.Close()

	var jiraResult IssueInfo

	if err := json.NewDecoder(res.Body).Decode(&jiraResult); err != nil {
		return nil, errors.Wrap(err, "json.NewDecoder(res.Body).Decode")
	}

	_ = cache.Client().Set(ctx, cacheKey, jiraResult, cache.MiddleTTL)

	return &jiraResult, nil
}
