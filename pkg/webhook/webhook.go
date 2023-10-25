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
package webhook

import (
	"context"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/maksim-paskal/kubernetes-manager/pkg/webhook/aws"
	"github.com/maksim-paskal/kubernetes-manager/pkg/webhook/azure"
	"github.com/pkg/errors"
)

// provider interface to process events.
type Provider interface {
	Init(webhook config.WebHook, webhookMessage types.WebhookMessage) error
	Process(ctx context.Context) error
}

// create new webbhok event.
func NewEvent(ctx context.Context, message types.WebhookMessage) error {
	for _, condition := range config.Get().WebHooks {
		if condition.Cluster == message.Cluster && condition.Namespace == message.Namespace {
			if err := processEvent(ctx, condition, message); err != nil {
				return errors.Wrap(err, "error while processing event")
			}
		}
	}

	return nil
}

// test valid webhooks config.
func CheckConfig() error {
	for _, condition := range config.Get().WebHooks {
		provider, err := NewProvider(condition.Provider)
		if err != nil {
			return errors.Wrap(err, "find valid provider")
		}

		// init provider with null message.
		if err := provider.Init(condition, types.WebhookMessage{}); err != nil {
			if err != nil {
				return errors.Wrap(err, "find valid provider")
			}
		}
	}

	return nil
}

// process event.
func processEvent(ctx context.Context, condition config.WebHook, message types.WebhookMessage) error {
	provider, err := NewProvider(condition.Provider)
	if err != nil {
		return errors.Wrap(err, "find valid provider")
	}

	if err := provider.Init(condition, message); err != nil {
		return errors.Wrap(err, "provider init error")
	}

	return errors.Wrap(provider.Process(ctx), "provider processing error")
}

func NewProvider(provider string) (Provider, error) { //nolint:ireturn
	switch provider {
	case "aws":
		return new(aws.Provider), nil
	case "azure":
		return new(azure.Provider), nil
	default:
		return nil, errors.New("no provider was found " + provider)
	}
}
