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
package sentry

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func NewClient(endpoint string) *Client {
	return &Client{
		Endpoint:   strings.TrimRight(endpoint, "/"),
		HTTPClient: http.DefaultClient,
	}
}

type Client struct {
	Endpoint     string
	Token        string
	Organization string
	HTTPClient   *http.Client
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Metadata struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Issue struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Culprit   string   `json:"culprit"`
	ShortID   string   `json:"shortId"`
	Level     string   `json:"level"`
	LastSeen  string   `json:"lastSeen"`
	Metadata  Metadata `json:"metadata"`
	PermaLink string   `json:"permalink"`
	Project   Project  `json:"project"`
	SubStatus string   `json:"substatus"`
}

func (i *Issue) LastSeenTime() time.Time {
	lastSeen, err := time.Parse(time.RFC3339, i.LastSeen)
	if err != nil {
		return time.Time{}
	}

	return lastSeen
}

func (c *Client) request(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.Token)

	log.Debug("request: " + req.URL.String())

	return req, nil
}

func (c *Client) GetProjectURL(slug string) string {
	return c.Endpoint + "/organizations/" + c.Organization + "/projects/" + slug + "/"
}

func (c *Client) GetIssuesExternalLink(query, period string) string {
	lintQuery := url.Values{
		"query":       []string{query},
		"statsPeriod": []string{period},
	}

	return c.Endpoint + "/organizations/" + c.Organization + "/issues/?" + lintQuery.Encode()
}

// Get issues from sentry
// (WARNING) token must have all teams.
func (c *Client) GetIssues(ctx context.Context, query, period, limit string) ([]Issue, error) {
	url, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing url")
	}

	url.Path = "/api/0/organizations/" + c.Organization + "/issues/"

	q := url.Query()
	q.Set("statsPeriod", period)
	q.Set("limit", limit)
	q.Set("query", query)

	url.RawQuery = q.Encode()

	req, err := c.request(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error sending request")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error response code: " + resp.Status)
	}

	defer resp.Body.Close()

	issues := make([]Issue, 0)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response")
	}

	log.Debug("response: " + string(respBody))

	if err := json.Unmarshal(respBody, &issues); err != nil {
		return nil, errors.Wrap(err, "error decoding response")
	}

	return issues, nil
}
