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

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type GetRemoteServerItemStatus string

const (
	GetRemoteServerItemStatusRunning GetRemoteServerItemStatus = "Running"
	GetRemoteServerItemStatusStoped  GetRemoteServerItemStatus = "Stoped"
)

type GetRemoteServerItem struct {
	Cloud  string
	ID     string
	Name   string
	Status GetRemoteServerItemStatus
	IPv4   string
	Labels map[string]string
}

var errNoHetznerCloudToken = errors.New("no hetzner cloud token, please set it in config file")

// return all remote servers.
func GetRemoteServers(ctx context.Context) ([]*GetRemoteServerItem, error) {
	if len(config.Get().RemoteServer.HetznerToken) == 0 {
		return nil, errNoHetznerCloudToken
	}

	hcloundClient := client.GetHcloudClient()

	result := make([]*GetRemoteServerItem, 0)

	servers, err := hcloundClient.Server.All(ctx)
	if err != nil {
		return result, errors.Wrap(err, "can not get servers")
	}

	for _, server := range servers {
		status := GetRemoteServerItemStatusStoped

		if server.Status == hcloud.ServerStatusRunning {
			status = GetRemoteServerItemStatusRunning
		}

		serverName := server.Name

		if owner, ok := server.Labels["owner"]; ok {
			serverName = owner
		}

		result = append(result, &GetRemoteServerItem{
			Cloud:  "hcloud",
			ID:     strconv.Itoa(server.ID),
			Name:   serverName,
			Status: status,
			IPv4:   server.PublicNet.IPv4.IP.String(),
			Labels: formattedRemoteServerLabels(server.Labels),
		})
	}

	return result, nil
}

func formattedRemoteServerLabels(labels map[string]string) map[string]string {
	formatted := make(map[string]string)

	formatUnixTime := func(v string) string {
		d, err := utils.UnixToTime(v)
		if err != nil {
			log.WithError(err).Errorf("error parsing time %s", v)

			return "error parsing time"
		}

		text := utils.HumanizeDuration(utils.HumanizeDurationShort, time.Since(d))

		return fmt.Sprintf("%s ago", text)
	}

	for k, v := range labels {
		if k == "lastPowerOnTime" {
			formatted[k] = formatUnixTime(v)
		}
	}

	return formatted
}
