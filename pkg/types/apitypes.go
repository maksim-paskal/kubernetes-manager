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
package types

// These types are shared by the API server (pkg/api, pkg/config) and the CLI.
// They must not import packages that pull in cloud or Kubernetes SDKs.

import (
	"bytes"
	"encoding/json"
	"slices"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
)

type Links struct {
	SentryURL     string
	SlackURL      string
	LogsURL       string
	LogsPodURL    string
	PhpMyAdminURL string
	MetricsURL    string
	TracingURL    string
	JiraURL       string
	Others        []OtherLink
}

func (l *Links) FormatedLinks(namespace string) (*Links, error) {
	linksJSON, err := json.Marshal(l)
	if err != nil {
		return nil, errors.Wrap(err, "error while json.Marshal")
	}

	linksJSONFormatted := linksJSON
	linksJSONFormatted = bytes.ReplaceAll(linksJSONFormatted, []byte("__Namespace__"), []byte(namespace))

	link := Links{}

	err = json.Unmarshal(linksJSONFormatted, &link)
	if err != nil {
		return nil, errors.Wrap(err, "error while json.Unmarshal")
	}

	return &link, nil
}

type OtherLink struct {
	Name        string
	URL         string
	Description string
}

type Template struct {
	NamespacePattern string // display template on some namespace
	Display          string
	Data             string
}

type WikiPage struct {
	Title     string
	ProjectID int
	Slug      string
	Size      int
}

type GetFrontConfigItem struct {
	ClusterName string
	Maintenance bool
	Links       *Links
}

type GetFrontConfigResult struct {
	Version        string
	Links          *Links
	Clusters       []*GetFrontConfigItem
	DebugTemplates []*Template
	WikiPages      []*WikiPage
}

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

type GetRemoteServerItemStatus string

const (
	LastPowerOnTimeLabel = "lastPowerOnTime"
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
		LastPowerOnTimeLabel,
	}

	return slices.Contains(validKeys, l.Key)
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
	Links           []*OtherLink
}

func (i *GetRemoteServerItem) GetLastPowerOnTime() (time.Time, error) {
	lastPowerOnTimeString, ok := i.Labels[LastPowerOnTimeLabel]
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

type GetGitlabProjectsItem struct {
	ProjectID      int64
	Name           string
	Description    string
	DefaultBranch  string
	WebURL         string
	TagsList       []string
	AdditionalInfo *GetGitlabProjectsInfoItem // custom field for front end
	Deploy         string                     // custom field for front end
	// custom field for front end
	RowVariant     string `json:"_rowVariant"` //nolint:tagliatelle
	Required       bool
	SelectedBranch string
}

type GetGitlabProjectsInfoItem struct {
	PodRunning     *GetPodByImageResult
	Pipelines      *GetGitlabPipelinesStatusResults
	CommitsBehind  *int
	DefaultBranch  *string
	BranchNotFound bool
}

type GetPodByImageResult struct {
	Tag     string
	GitHash string
	Found   bool
}

type GetGitlabPipelinesStatusResults struct {
	LastErrorPipeline   string
	LastRunningPipeline string
	LastSuccessPipeline string
}

// StartNewEnvironmentInput is the CLI copy of the request, the server type adds validation on top of it.
type StartNewEnvironmentInput struct {
	Profile  string
	Services string
	Cluster  string
	Name     string
}

// Environment is the CLI copy of the environment response, it holds the exported fields only.
type Environment struct {
	ID                      string
	Cluster                 string
	Namespace               string
	NamespaceStatus         string
	NamespaceCreated        string
	NamespaceCreatedDays    int
	NamespaceCreatedBy      string
	NamespaceLastScaled     string
	NamespaceLastScaledDays int
	NamespaceAnnotations    map[string]string
	NamespaceLabels         map[string]string
	LikedByMe               bool
	NamespaceDescription    string
	Links                   *Links
	Hosts                   []string
	HostsInternal           []string
	PodsInfo                *PodsInfo
	NamespaceBadges         []*EnvironmentBadge
}

type PodsInfo struct {
	PodsTotal       int64
	PodsReady       int64
	PodsFailed      int64
	PodsStarting    int64
	PodsTerminating int64
	PodsFailedName  []string

	CPURequests     string
	MemoryRequests  string
	StorageRequests string
}

type EnvironmentBadge struct {
	Key         string
	Value       string
	Description string
}
