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
package cache

import (
	"context"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/cache/noop"
	"github.com/maksim-paskal/kubernetes-manager/pkg/cache/redis"
	"github.com/pkg/errors"
)

const (
	MaxTTL  = 0
	HighTTL = 24 * time.Hour
	LowTTL  = 10 * time.Minute
)

type Provider interface {
	Get(ctx context.Context, key string, value any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}

type ProviderName string

const (
	NoopProvider  ProviderName = "noop"
	RedisProvider ProviderName = "redis"
)

var provider Provider

func Init(ctx context.Context, providerName ProviderName, providerConfig interface{}) error {
	switch providerName {
	case NoopProvider:
		provider = &noop.Provider{}
	case RedisProvider:
		redisProvider, err := redis.NewProvider(ctx, providerConfig)
		if err != nil {
			return errors.Wrap(err, "can not create redis provider")
		}

		provider = redisProvider
	default:
		return errors.Errorf("unknown cache provider %s", providerName)
	}

	return nil
}

func Client() Provider { //nolint:ireturn
	return provider
}
