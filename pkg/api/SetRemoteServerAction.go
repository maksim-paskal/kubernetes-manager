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
	"fmt"
	"strconv"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
)

type SetRemoteServerStatusAction string

const (
	SetRemoteServerStatusPowerOn  SetRemoteServerStatusAction = "PowerOn"
	SetRemoteServerStatusPowerOff SetRemoteServerStatusAction = "PowerOff"
)

func (a SetRemoteServerStatusAction) Validate() error {
	if a == SetRemoteServerStatusPowerOn {
		return nil
	}

	if a == SetRemoteServerStatusPowerOff {
		return nil
	}

	return errors.New("unknown status")
}

type SetRemoteServerActionInput struct {
	Cloud  string
	ID     string
	Action SetRemoteServerStatusAction
}

// return all remote servers.
func SetRemoteServerAction(ctx context.Context, input SetRemoteServerActionInput) error {
	if input.Cloud != "hcloud" {
		return errors.New("cloud not supported")
	}

	if err := input.Action.Validate(); err != nil {
		return errors.Wrapf(err, "error validate action %s", input.Action)
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

	if input.Action == SetRemoteServerStatusPowerOff {
		_, _, err = hcloundClient.Server.Poweroff(ctx, server)
		if err != nil {
			return errors.Wrap(err, "can power off server")
		}
	}

	if input.Action == SetRemoteServerStatusPowerOn {
		_, _, err = hcloundClient.Server.Poweron(ctx, server)
		if err != nil {
			return errors.Wrap(err, "can power on server")
		}
	}

	labels := map[string]string{
		fmt.Sprintf("last%sTime", string(input.Action)): utils.TimeToUnix(time.Now()),
	}

	err = SetRemoteServerLabels(ctx, server, labels)
	if err != nil {
		return errors.Wrap(err, "error updating server")
	}

	return nil
}
