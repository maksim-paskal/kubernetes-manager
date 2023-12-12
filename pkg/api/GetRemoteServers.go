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
	"strings"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type GetRemoteServerItemStatus string

const (
	lastPowerOnTimeLabel = "lastPowerOnTime"
	staledTime           = 14 * 24 * time.Hour
)

const (
	GetRemoteServerItemStatusRunning GetRemoteServerItemStatus = "Running"
	GetRemoteServerItemStatusStoped  GetRemoteServerItemStatus = "Stoped"
)

type GetRemoteServerLabel struct {
	Key         string
	Value       string
	Description string
}

func (l *GetRemoteServerLabel) ValidKey() bool {
	validKeys := []string{
		lastPowerOnTimeLabel,
	}

	for _, key := range validKeys {
		if l.Key == key {
			return true
		}
	}

	return false
}

type GetRemoteServerItem struct {
	Cloud           string
	ID              string
	Name            string
	Status          GetRemoteServerItemStatus
	IPv4            string
	Created         time.Time
	Labels          map[string]string
	FormattedLabels []*GetRemoteServerLabel
	Links           []*config.OtherLink
}

func (i *GetRemoteServerItem) GetLastPowerOnTime() (time.Time, error) {
	lastPowerOnTimeString, ok := i.Labels[lastPowerOnTimeLabel]
	if !ok {
		return i.Created, nil
	}

	return utils.UnixToTime(lastPowerOnTimeString)
}

func (i *GetRemoteServerItem) IsStaled() bool {
	lastPowerOnTime, err := i.GetLastPowerOnTime()
	if err != nil {
		return false
	}

	return time.Since(lastPowerOnTime) > staledTime
}

var errNoHetznerCloudToken = errors.New("no hetzner cloud token, please set it in config file")

// return all remote servers.
func GetRemoteServers(ctx context.Context) ([]*GetRemoteServerItem, error) {
	ctx, span := telemetry.Start(ctx, "api.GetRemoteServers")
	defer span.End()

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

		item := &GetRemoteServerItem{
			Cloud:           "hcloud",
			ID:              strconv.Itoa(server.ID),
			Name:            serverName,
			Status:          status,
			Created:         server.Created,
			IPv4:            server.PublicNet.IPv4.IP.String(),
			Labels:          server.Labels,
			FormattedLabels: getRemoteServerLabels(server.Labels),
		}

		links := make([]*config.OtherLink, len(config.Get().RemoteServer.Links))
		for id, link := range config.Get().RemoteServer.Links {
			links[id] = &config.OtherLink{
				Name:        link.Name,
				Description: link.Description,
			}

			urlFormatted, err := utils.GetTemplatedResult(ctx, link.URL, item)
			if err != nil {
				log.WithError(err).Errorf("error parsing link %s", link.URL)
				links[id].URL = link.URL
			} else {
				links[id].URL = string(urlFormatted)
			}
		}

		item.Links = links

		if item.IsStaled() {
			item.FormattedLabels = append(item.FormattedLabels, &GetRemoteServerLabel{
				Key:         "staled",
				Value:       "true",
				Description: "server is staled",
			})
		}

		result = append(result, item)
	}

	return result, nil
}

func getRemoteServerLabels(labels map[string]string) []*GetRemoteServerLabel {
	result := make([]*GetRemoteServerLabel, 0)

	formatUnixTime := func(v string) (string, string) {
		d, err := utils.UnixToTime(v)
		if err != nil {
			log.WithError(err).Errorf("error parsing time %s", v)

			return "error parsing time", ""
		}

		text := utils.HumanizeDuration(utils.HumanizeDurationShort, time.Since(d))

		return fmt.Sprintf("%s ago", text), utils.TimeToString(d)
	}

	for k, v := range labels {
		item := &GetRemoteServerLabel{
			Key: k,
		}

		if !item.ValidKey() {
			continue
		}

		if strings.HasSuffix(k, "Time") {
			item.Value, item.Description = formatUnixTime(v)
		}

		if item.Key == lastPowerOnTimeLabel {
			item.Key = badgeLastStarted
		}

		result = append(result, item)
	}

	return result
}
