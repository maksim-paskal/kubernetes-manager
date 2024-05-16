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
package httpcall

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type ProviderConfig struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    string
	Timeout time.Duration
}

type Provider struct {
	Config    ProviderConfig
	Condition config.WebHook
	Message   types.WebhookMessage
}

var httpClient = &http.Client{
	Jar: nil,
}

func (provider *Provider) Init(condition config.WebHook, message types.WebhookMessage) error {
	log.Info("init notify provider")

	configBytes, err := json.Marshal(condition.Config)
	if err != nil {
		return errors.Wrap(err, "invalid condition config")
	}

	err = json.Unmarshal(configBytes, &provider.Config)
	if err != nil {
		return errors.Wrap(err, "invalid config")
	}

	if provider.Config.Timeout == 0 {
		provider.Config.Timeout = 30 * time.Second //nolint:gomnd,mnd
	}

	if len(provider.Config.Method) == 0 {
		provider.Config.Method = http.MethodPost
	}

	provider.Condition = condition
	provider.Message = message

	return nil
}

func (provider *Provider) Process(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "httpcall.Process")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, provider.Config.Timeout)
	defer cancel()

	log.Info("process notify provider")

	result, err := utils.GetTemplatedResult(ctx, provider.Config.Body, provider)
	if err != nil {
		return errors.Wrap(err, "error templating body")
	}

	req, err := http.NewRequestWithContext(ctx,
		provider.Config.Method,
		provider.Config.URL,
		bytes.NewBuffer(result),
	)
	if err != nil {
		return errors.Wrap(err, "error creating request")
	}

	for key, value := range provider.Config.Headers {
		req.Header.Set(key, value)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error sending request")
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("error sending request, status code: %d", res.StatusCode)
	}

	return nil
}
