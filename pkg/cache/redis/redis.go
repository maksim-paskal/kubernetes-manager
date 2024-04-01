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
package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/metrics"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type Provider struct {
	client *redis.Client
}

type ProviderConfig struct {
	URL       string
	TLSConfig tls.Config
}

func NewProvider(ctx context.Context, config interface{}) (*Provider, error) {
	providerConfig := ProviderConfig{}

	if config != nil {
		configBytes, err := json.Marshal(config)
		if err != nil {
			return nil, errors.Wrap(err, "json.Marshal")
		}

		if err := json.Unmarshal(configBytes, &providerConfig); err != nil {
			return nil, errors.Wrap(err, "json.Unmarshal")
		}
	}

	if len(providerConfig.URL) == 0 {
		return nil, errors.New("missing URL")
	}

	opt, err := redis.ParseURL(providerConfig.URL)
	if err != nil {
		return nil, errors.Wrap(err, "redis.ParseURL")
	}

	opt.TLSConfig = &providerConfig.TLSConfig

	provider := &Provider{
		client: redis.NewClient(opt),
	}

	if err := provider.client.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrap(err, "provider.client.Ping")
	}

	return provider, nil
}

func (p *Provider) Get(ctx context.Context, key string, value any) error {
	ctx, span := telemetry.Start(ctx, "cache.redis.Get")
	defer span.End()

	val, err := p.client.Get(ctx, key).Bytes()
	if err != nil {
		return errors.Wrap(err, "p.client.Get")
	}

	if len(val) == 0 {
		return nil
	}

	err = json.Unmarshal(val, &value)
	if err != nil {
		return errors.Wrap(err, "json.Unmarshal")
	}

	return nil
}

func (p *Provider) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	ctx, span := telemetry.Start(ctx, "cache.redis.Set")
	defer span.End()

	valueBytes, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	if err := p.client.Set(ctx, key, string(valueBytes), ttl).Err(); err != nil {
		return errors.Wrap(err, "p.client.Set")
	}

	metrics.CacheAdd.Inc()

	return nil
}
