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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	mathRand "math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/briandowns/spinner"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	api_user "github.com/maksim-paskal/kubernetes-manager/pkg/user"
	"github.com/pkg/errors"
)

func NewClient() *Client {
	c := &Client{
		// write spinner to stderr, stdout must stay clean for --json output
		s: spinner.New(spinner.CharSets[7], 100*time.Millisecond, spinner.WithWriterFile(os.Stderr)),
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}

	// PreUpdate runs on every frame in the goroutine of the spinner, the dots advance every 5 frames
	// (about half a second) through ".", ".." and "...". It must not take the spinner lock, Stop is called under it.
	frame := 0

	c.s.PreUpdate = func(s *spinner.Spinner) {
		text, _ := c.spinnerText.Load().(string)

		s.Suffix = text + strings.Repeat(".", frame/5%3+1)
		frame++
	}

	return c
}

type BasicAuth struct {
	Login    string
	Password string
}

type Client struct {
	s *spinner.Spinner

	// spinnerUsers is the number of requests that show the spinner, spinnerText is the text of the last one,
	// spinnerStoppedAt is when the spinner stopped the last time
	spinnerMu        sync.Mutex
	spinnerUsers     int
	spinnerText      atomic.Value
	spinnerStoppedAt time.Time

	HTTPClient *http.Client

	BaseURL string
	Auth    *BasicAuth

	// DevUser is sent as the X-Auth-Request-User header, set only for local development
	DevUser string

	// PollInterval is how often the Wait methods check the environment, zero means the default
	PollInterval time.Duration
}

func (c *Client) WithBaseURL(baseURL string) *Client {
	c.BaseURL = baseURL
	c.BaseURL = strings.TrimSpace(c.BaseURL)
	c.BaseURL = strings.TrimRight(c.BaseURL, "/")

	return c
}

func (c *Client) WithBasicAuth(login, password string) *Client {
	c.Auth = &BasicAuth{
		Login:    login,
		Password: password,
	}

	return c
}

func (c *Client) WithDevUser(user string) *Client {
	c.DevUser = user

	return c
}

func (c *Client) Validate() error {
	if c.BaseURL == "" {
		return errors.New("base URL is required")
	}

	if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
		return errors.New("base URL must start with http:// or https://")
	}

	if c.Auth == nil || c.Auth.Login == "" || c.Auth.Password == "" {
		return errors.New("both login and password are required")
	}

	return nil
}

func (c *Client) post(ctx context.Context, name string, input any) ([]byte, error) {
	buf, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	if slog.Default().Enabled(ctx, slog.LevelDebug) {
		slog.DebugContext(ctx, debugBody("request body", buf))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+name, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	resp, err := c.call(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d\n%s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) get[T any](ctx context.Context, name string, params map[string]string) (T, error) {
	var zero T

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+name, nil)
	if err != nil {
		return zero, err
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}

	req.URL.RawQuery = q.Encode()

	resp, err := c.call(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return zero, fmt.Errorf("unexpected status code: %d\n%s", resp.StatusCode, string(body))
	}

	// raw text responses (e.g. kubeconfig) are not wrapped in an envelope
	if _, ok := any(zero).(string); ok {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return zero, err
		}

		result, ok := any(string(body)).(T)
		if !ok {
			return zero, fmt.Errorf("cannot convert raw response to %T", zero)
		}

		return result, nil
	}

	var envelope struct {
		Result T
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return zero, errors.Wrap(err, "failed to decode envelope")
	}

	return envelope.Result, nil
}

// secretHeaders carry credentials, their values are never logged.
var secretHeaders = []string{"Authorization", "Cookie", "Set-Cookie"}

func redactHeaders(h http.Header) http.Header {
	out := h.Clone()

	for _, name := range secretHeaders {
		if out.Get(name) != "" {
			out.Set(name, "REDACTED")
		}
	}

	return out
}

// debugBody returns the debug message with the body below it. JSON is indented, it is put in the message
// because the log handler quotes and escapes attribute values, which makes a body unreadable.
func debugBody(message string, body []byte) string {
	body = bytes.TrimSpace(body)

	var out bytes.Buffer

	if json.Indent(&out, body, "", "  ") != nil {
		out.Reset()
		out.Write(body)
	}

	return message + "\n" + out.String()
}

// spinnerVerbs are shown at random while a request is in flight.
var spinnerVerbs = []string{"Loading", "Thinking", "Pondering", "Cogitating", "Mulling", "Noodling"}

// spinnerDelay is how long a request runs before the spinner shows, to avoid flickering for fast requests.
// A request that starts while the spinner shows, or within the delay after it stopped, shows it at once, so the
// requests of one command do not make the spinner blink.
const spinnerDelay = time.Second

// showSpinner shows the spinner when the request is still running after a second, the returned func stops it.
// The requests can run in parallel, the spinner is shown while any of them is in flight.
func (c *Client) showSpinner(text string) func() {
	// the lock keeps a timer that fires while the request finishes from starting the spinner after it was stopped
	var (
		mu       sync.Mutex
		stopped  bool
		starting bool
	)

	start := func() {
		mu.Lock()
		defer mu.Unlock()

		if stopped {
			return
		}

		starting = true

		c.spinnerMu.Lock()
		defer c.spinnerMu.Unlock()

		c.spinnerText.Store(text)

		if c.spinnerUsers == 0 {
			c.s.Start()
		}

		c.spinnerUsers++
	}

	var timer *time.Timer

	c.spinnerMu.Lock()
	visible := c.spinnerUsers > 0 || time.Since(c.spinnerStoppedAt) < spinnerDelay
	c.spinnerMu.Unlock()

	if visible {
		start()
	} else {
		timer = time.AfterFunc(spinnerDelay, start)
	}

	return func() {
		mu.Lock()
		defer mu.Unlock()

		stopped = true

		if timer != nil {
			timer.Stop()
		}

		if !starting {
			return
		}

		c.spinnerMu.Lock()
		defer c.spinnerMu.Unlock()

		if c.spinnerUsers--; c.spinnerUsers == 0 {
			c.s.Stop()
			c.spinnerStoppedAt = time.Now()
		}
	}
}

func (c *Client) call(req *http.Request) (*http.Response, error) {
	text := " " + spinnerVerbs[mathRand.Intn(len(spinnerVerbs))] + " " + strings.TrimPrefix(req.URL.Path, "/api/") //nolint:gosec

	defer c.showSpinner(text)()

	req.SetBasicAuth(c.Auth.Login, c.Auth.Password)

	if c.DevUser != "" {
		slog.Debug("Setting development user header", "user", c.DevUser)
		req.Header.Set("X-Auth-Request-User", c.DevUser)
	}

	status := 0

	var responseHeaders http.Header

	startTime := time.Now()

	defer func() {
		slog.Debug("Making request", "url", req.URL.String(), "method", req.Method, "status", status,
			"headers", redactHeaders(req.Header), "responseHeaders", redactHeaders(responseHeaders), "duration", time.Since(startTime))
	}()

	// the timeout of the HTTP client is a default for the requests without a deadline, a deadline of the command
	// (--timeout, --wait.timeout) is the limit the user chose, so a server operation may take longer than a minute
	httpClient := c.HTTPClient

	if _, ok := req.Context().Deadline(); ok {
		withDeadline := *c.HTTPClient
		withDeadline.Timeout = 0
		httpClient = &withDeadline
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	status = resp.StatusCode
	responseHeaders = resp.Header

	// kubeconfig contains a service account token, its body is never logged
	if slog.Default().Enabled(req.Context(), slog.LevelDebug) && !strings.HasSuffix(req.URL.Path, "/kubeconfig") {
		b, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body")
		}

		slog.DebugContext(req.Context(), debugBody("response body", b))

		resp.Body = io.NopCloser(bytes.NewReader(b))
	}

	return resp, nil
}

func (c *Client) GetUser(ctx context.Context) (*api_user.User, error) {
	result, err := c.get[api_user.User](ctx, "/api/whoami", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	return &result, nil
}

type RemoteServer types.GetRemoteServerItem

func (c *Client) GetRemoteServers(ctx context.Context, all bool) ([]RemoteServer, error) {
	params := map[string]string{}
	params["all"] = strconv.FormatBool(all)

	result, err := c.get[[]RemoteServer](ctx, "/api/remote-servers", params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get remote servers")
	}

	return result, nil
}

func (c *Client) GetRemoteServer(ctx context.Context, id string) (*RemoteServer, error) {
	servers, err := c.GetRemoteServers(ctx, true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get remote servers")
	}

	for _, server := range servers {
		if server.ID == id {
			return &server, nil
		}
	}

	return nil, errors.New("remote server not found")
}

func (c *Client) RemoteServerAction(ctx context.Context, input types.SetRemoteServerActionInput) ([]byte, error) {
	return c.post(ctx, "/api/make-remote-server-action", input)
}

type Environment types.Environment

// ValidateEnvironmentID checks that id is <cluster>:<namespace>, the server checks that the environment exists.
func ValidateEnvironmentID(id string) error {
	if _, err := types.NewIDInfo(id); err != nil {
		return errors.Wrapf(err, "invalid environment id %q, expected <cluster>:<namespace>", id)
	}

	return nil
}

// envPath returns the API path of an operation on the environment.
func envPath(id, op string) (string, error) {
	if err := ValidateEnvironmentID(id); err != nil {
		return "", err
	}

	return "/api/" + url.PathEscape(id) + "/" + op, nil
}

func (c *Client) GetEnvironments(ctx context.Context, params map[string]string) ([]Environment, error) {
	result, err := c.get[[]Environment](ctx, "/api/environments", params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get environments")
	}

	return result, nil
}

func (c *Client) GetEnvironment(ctx context.Context, id string) (*Environment, error) {
	path, err := envPath(id, "info")
	if err != nil {
		return nil, err
	}

	result, err := c.get[Environment](ctx, path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get environment")
	}

	return &result, nil
}

// ServiceRef is a service of an environment: the ref (branch or tag) to deploy from the project.
type ServiceRef struct {
	ProjectID int
	Ref       string
}

// ParseServiceRef parses <projectID>:<ref>, the ref itself can not contain ':' or ';'.
func ParseServiceRef(value string) (ServiceRef, error) {
	projectID, ref, found := strings.Cut(value, ":")
	if !found || ref == "" || strings.ContainsAny(ref, ":;") {
		return ServiceRef{}, errors.Errorf("invalid service %q, expected <projectID>:<ref>", value)
	}

	id, err := strconv.Atoi(projectID)
	if err != nil || id <= 0 {
		return ServiceRef{}, errors.Errorf("invalid service %q, project id must be a positive number", value)
	}

	return ServiceRef{ProjectID: id, Ref: ref}, nil
}

func (s ServiceRef) String() string {
	return strconv.Itoa(s.ProjectID) + ":" + s.Ref
}

// formatServiceRefs returns the services in the format of the server, items are separated by ';'.
func formatServiceRefs(services []ServiceRef) string {
	items := make([]string, len(services))

	for i, service := range services {
		items[i] = service.String()
	}

	return strings.Join(items, ";")
}

// CreateEnvironmentInput is the request to create an environment from a profile.
type CreateEnvironmentInput struct {
	Profile  string
	Cluster  string
	Name     string
	Services []ServiceRef
}

// CreateEnvironment creates a new environment and returns its ID.
func (c *Client) CreateEnvironment(ctx context.Context, input CreateEnvironmentInput) (string, error) {
	body, err := c.post(ctx, "/api/make-create-environment", types.StartNewEnvironmentInput{
		Profile:  input.Profile,
		Cluster:  input.Cluster,
		Name:     input.Name,
		Services: formatServiceRefs(input.Services),
	})
	if err != nil {
		return "", err
	}

	var envelope struct {
		Result string
	}

	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", errors.Wrap(err, "failed to decode envelope")
	}

	return envelope.Result, nil
}

type Profile struct {
	Name  string
	Value string
}

func (c *Client) GetProfiles(ctx context.Context) ([]Profile, error) {
	result, err := c.get[[]Profile](ctx, "/api/project-profiles", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get profiles")
	}

	return result, nil
}

type Service types.GetGitlabProjectsItem

func (c *Client) GetServices(ctx context.Context, profile string) ([]Service, error) {
	result, err := c.get[[]Service](ctx, "/api/external-services", map[string]string{"profile": profile})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get services")
	}

	return result, nil
}

func (c *Client) GetRef(ctx context.Context, id string) ([]string, error) {
	result, err := c.get[[]string](ctx, "/api/project-refs", map[string]string{"id": id})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get refs")
	}

	return result, nil
}

func (c *Client) DeleteEnvironment(ctx context.Context, id string) error {
	path, err := envPath(id, "make-delete")
	if err != nil {
		return err
	}

	_, err = c.post(ctx, path, map[string]string{"id": id})
	if err != nil {
		return errors.Wrap(err, "failed to delete environment")
	}

	return nil
}

// StartEnvironment starts all pods in the environment, the server scales it up in background.
func (c *Client) StartEnvironment(ctx context.Context, id string) error {
	path, err := envPath(id, "make-start")
	if err != nil {
		return err
	}

	_, err = c.post(ctx, path, map[string]string{"id": id})
	if err != nil {
		return errors.Wrap(err, "failed to start environment")
	}

	return nil
}

// PauseEnvironment pauses all pods in the environment, the server scales it down in background.
func (c *Client) PauseEnvironment(ctx context.Context, id string) error {
	path, err := envPath(id, "make-pause")
	if err != nil {
		return err
	}

	_, err = c.post(ctx, path, map[string]string{"id": id})
	if err != nil {
		return errors.Wrap(err, "failed to pause environment")
	}

	return nil
}

// DisableMTLSEnvironment disables mTLS in the environment, the server updates the envoy control plane synchronously.
func (c *Client) DisableMTLSEnvironment(ctx context.Context, id string) error {
	path, err := envPath(id, "make-disable-mtls")
	if err != nil {
		return err
	}

	_, err = c.post(ctx, path, map[string]string{"id": id})
	if err != nil {
		return errors.Wrap(err, "failed to disable mTLS in environment")
	}

	return nil
}

// ScaleDownDelayEnvironment delays the automatic pause of the environment.
func (c *Client) ScaleDownDelayEnvironment(ctx context.Context, id string, delay time.Duration) error {
	path, err := envPath(id, "make-scaledown-delay")
	if err != nil {
		return err
	}

	_, err = c.post(ctx, path, map[string]string{"Delay": delay.String()})
	if err != nil {
		return errors.Wrap(err, "failed to delay scaledown of environment")
	}

	return nil
}

// GetIssues returns the Sentry issues of the environment.
func (c *Client) GetIssues(ctx context.Context, id string) (*types.GetIssuesResponse, error) {
	path, err := envPath(id, "issues")
	if err != nil {
		return nil, err
	}

	result, err := c.get[types.GetIssuesResponse](ctx, path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get issues")
	}

	return &result, nil
}

func (c *Client) GetConfig(ctx context.Context) (*types.GetFrontConfigResult, error) {
	result, err := c.get[types.GetFrontConfigResult](ctx, "/api/front-config", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config")
	}

	return &result, nil
}

func (c *Client) GetKubeConfig(ctx context.Context, id string) (string, error) {
	path, err := envPath(id, "kubeconfig")
	if err != nil {
		return "", err
	}

	result, err := c.get[string](ctx, path, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to get config")
	}

	return result, nil
}

// GetAutotests returns the autotest runs of the environment, size 0 returns all of them.
func (c *Client) GetAutotests(ctx context.Context, id string, size int) ([]*types.AutotestPipeline, error) {
	details, err := c.getAutotestDetails(ctx, id, size)
	if err != nil {
		return nil, err
	}

	return details.Pipelines, nil
}

// GetAutotestActions returns the tests that can be started in the environment.
func (c *Client) GetAutotestActions(ctx context.Context, id string) ([]*types.AutotestAction, error) {
	// the runs are not needed, ask for the smallest page
	details, err := c.getAutotestDetails(ctx, id, 1)
	if err != nil {
		return nil, err
	}

	return details.Actions, nil
}

func (c *Client) getAutotestDetails(ctx context.Context, id string, size int) (*types.AutotestDetails, error) {
	path, err := envPath(id, "autotests")
	if err != nil {
		return nil, err
	}

	result, err := c.get[types.AutotestDetails](ctx, path, map[string]string{"size": strconv.Itoa(size)})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get autotests")
	}

	return &result, nil
}

// StopAutotest stops the autotest run with the pipeline id in the environment.
func (c *Client) StopAutotest(ctx context.Context, id, pipelineID string) error {
	path, err := envPath(id, "make-stop-autotest")
	if err != nil {
		return err
	}

	_, err = c.post(ctx, path, map[string]string{"PipelineID": pipelineID})
	if err != nil {
		return errors.Wrap(err, "failed to stop autotest")
	}

	return nil
}

// StartAutotest starts a run of the autotest in the environment and returns the message of the server.
func (c *Client) StartAutotest(ctx context.Context, id string, input types.StartAutotestInput) (string, error) {
	path, err := envPath(id, "make-start-autotest")
	if err != nil {
		return "", err
	}

	body, err := c.post(ctx, path, input)
	if err != nil {
		return "", errors.Wrap(err, "failed to start autotest")
	}

	var envelope struct {
		Result string
	}

	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", errors.Wrap(err, "failed to decode envelope")
	}

	return envelope.Result, nil
}
