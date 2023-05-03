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
	"strconv"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
)

type SetRemoteServerDelayInput struct {
	Cloud    string
	ID       string
	Duration string
}

// set autopause date for remote server.
func SetRemoteServerDelay(ctx context.Context, input SetRemoteServerDelayInput) error {
	if input.Cloud != "hcloud" {
		return errors.New("cloud not supported")
	}

	duration, err := time.ParseDuration(input.Duration)
	if err != nil {
		return errors.New("error parse duration")
	}

	hcloundClient := client.GetHcloudClient()

	id, err := strconv.Atoi(input.ID)
	if err != nil {
		return errors.New("can not parse id")
	}

	server, _, err := hcloundClient.Server.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "can not get server")
	}

	labels := map[string]string{
		config.LabelScaleDownDelayShort: utils.TimeToUnix(time.Now().Add(duration)),
	}

	err = SetRemoteServerLabels(ctx, server, labels)
	if err != nil {
		return errors.Wrap(err, "error set labels")
	}

	return nil
}
