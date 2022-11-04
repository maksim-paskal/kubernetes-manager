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
package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	defaultPort                      = 9000
	defaultRemoveBranchLastScaleDate = 10
	defaultBatchShedulePeriod        = 30 * time.Minute
	defaultBatchTimezone             = "UTC"

	ScaleDownHourMinPeriod = 19
	ScaleDownHourMaxPeriod = 5

	TemporaryTokenRandLength    = 5
	TemporaryTokenDurationHours = 10

	StaledNewNamespaceDurationDays = 3

	HoursInDay     = 24
	KeyValueLength = 2

	TrueValue = "true"

	Namespace             = "kubernetes-manager"
	FilterLabels          = Namespace + "=true"
	LabelType             = Namespace + "/type"
	LabelScaleDownDelay   = Namespace + "/scaleDownDelay"
	LabelLastScaleDate    = Namespace + "/lastScaleDate"
	LabelGitBranch        = Namespace + "/git-branch"
	LabelGitProjectID     = Namespace + "/git-project-id"
	LabelGitProjectOrigin = Namespace + "/git-project-origin"
	LabelRegistryTag      = Namespace + "/registry-tag"
	LabelSystemNamespace  = Namespace + "/system-namespace"
	TagNamespace          = Namespace + "/namespace"
	TagCluster            = Namespace + "/cluster"
	LabelNamespaceCreator = Namespace + "/user-creator"
	LabelInstalledProject = Namespace + "/project"
	LabelEnvironmentName  = Namespace + "/environment-name"
	LabelUserLiked        = Namespace + "/user-liked"
	LabelGitSyncOrigin    = Namespace + "/git-sync-origin"
	LabelGitSyncBranch    = Namespace + "/git-sync-branch"
)

type Links struct {
	SentryURL     string      `yaml:"sentryUrl"`
	SlackURL      string      `yaml:"slackUrl"`
	LogsURL       string      `yaml:"logsUrl"`
	LogsPodURL    string      `yaml:"logsPodUrl"`
	PhpMyAdminURL string      `yaml:"phpMyAdminUrl"`
	MetricsURL    string      `yaml:"metricsUrl"`
	TracingURL    string      `yaml:"tracingUrl"`
	Others        []OtherLink `yaml:"others"`
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
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Template struct {
	NamespacePattern string // display template on some namespace
	Display          string
	Data             string
}

type KubernetesEndpoint struct {
	Name             string
	KubeConfigPath   string
	KubeConfigServer string
	Links            *Links
}

type ProjectProfile struct {
	Name            string
	NamespacePrefix string
	DefaultBranch   string // use default branch for project (comma separated format projectId=main)
	Required        string // project ids to be required (comma separated)
	Exclude         string // project ids to exclude (comma separated)
	Include         string // project ids to include (comma separated)
}

func (p *ProjectProfile) GetRequired() []string {
	if len(p.Required) == 0 {
		return []string{}
	}

	return strings.Split(p.Required, ",")
}

func (p *ProjectProfile) GetExclude() []string {
	if len(p.Exclude) == 0 {
		return []string{}
	}

	return strings.Split(p.Exclude, ",")
}

func (p *ProjectProfile) GetInclude() []string {
	if len(p.Include) == 0 {
		return []string{}
	}

	return strings.Split(p.Include, ",")
}

func (p *ProjectProfile) IsProjectRequired(projectID int) bool {
	return utils.StringInSlice(strconv.Itoa(projectID), p.GetRequired())
}

func (p *ProjectProfile) GetProjectSelectedBranch(projectID int) string {
	if len(p.DefaultBranch) == 0 {
		return ""
	}

	for _, defaultBranch := range strings.Split(p.DefaultBranch, ",") {
		defaultBranchData := strings.Split(defaultBranch, "=")
		if len(defaultBranchData) != KeyValueLength {
			log.Errorf("invalid defaultBranch format %s", defaultBranch)

			continue
		}

		if defaultBranchData[0] == strconv.Itoa(projectID) {
			return defaultBranchData[1]
		}
	}

	return ""
}

type NamespaceMeta struct {
	Labels      map[string]string
	Annotations map[string]string
}

type WebHook struct {
	Provider  string
	Config    interface{}
	Cluster   string
	Namespace string
}

type Snapshot struct {
	ProjectID string
	Ref       string
}

type RemoteServer struct {
	HetznerToken string
}

//nolint:gochecknoglobals
var config = Type{
	ConfigPath: flag.String("config", os.Getenv("CONFIG"), "config"),
	LogLevel:   flag.String("log.level", "INFO", "logging level"),

	KubernetesEndpoints: []*KubernetesEndpoint{{
		Name:             "default",
		KubeConfigServer: GetEnvDefault("DEFAULT_CONFIG_SERVER", "https://cluster-public"),
		KubeConfigPath:   "",
	}},

	Port:      flag.Int("server.port", defaultPort, ""),
	FrontDist: flag.String("front.dist", "front/dist", ""),

	PodName:      flag.String("pod.name", os.Getenv("POD_NAME"), ""),
	PodNamespace: flag.String("pod.namespace", os.Getenv("POD_NAMESPACE"), ""),

	BatchShedulePeriod:   flag.Duration("batch.period", defaultBatchShedulePeriod, "batch shedule period"),
	BatchSheduleTimezone: flag.String("batch.timeZone", defaultBatchTimezone, "batch shedule timezone"),

	GitlabToken:     flag.String("gitlab.token", os.Getenv("GITLAB_TOKEN"), ""),
	GitlabTokenUser: flag.String("gitlab.token.user", os.Getenv("GITLAB_TOKEN_USER"), "username of token user (need to filter pipelines)"), //nolint:lll
	GitlabURL:       flag.String("gitlab.url", os.Getenv("GITLAB_URL"), ""),

	IngressHostDefaultProtocol: flag.String("ingress.show-protocol", "https", ""),

	RemoveBranchLastScaleDate: flag.Int("batch.removeBranchLastScaleDate", defaultRemoveBranchLastScaleDate, ""),

	ExternalServicesTopic: flag.String("externalServicesTopic", GetEnvDefault("EXTERNAL_SERVICES_TOPIC", "kubernetes-manager"), ""), //nolint:lll
	BatchEnabled:          flag.Bool("batch.enabled", true, "enable batch operations"),
}

type Type struct {
	ConfigPath                 *string `yaml:"configPath"`
	LogLevel                   *string `yaml:"logLevel"`
	Links                      *Links  `yaml:"links"`
	BatchEnabled               *bool
	NamespaceMeta              *NamespaceMeta
	DebugTemplates             []*Template
	ExternalServicesTemplates  []*Template
	ProjectProfiles            []*ProjectProfile
	KubernetesEndpoints        []*KubernetesEndpoint
	Port                       *int           `yaml:"port"`
	FrontDist                  *string        `yaml:"frontDist"`
	RemoveBranchDaysInactive   *int           `yaml:"removeBranchDaysInactive"`
	GitlabToken                *string        `yaml:"gitlabToken"`
	GitlabTokenUser            *string        `yaml:"gitlabTokenUser"`
	GitlabURL                  *string        `yaml:"gitlabUrl"`
	IngressHostDefaultProtocol *string        `yaml:"ingressHostDefaultProtocol"`
	RemoveBranchLastScaleDate  *int           `yaml:"removeBranchLastScaleDate"`
	ExternalServicesTopic      *string        `yaml:"externalServicesTopic"`
	BatchShedulePeriod         *time.Duration `yaml:"batchShedulePeriod"`
	BatchSheduleTimezone       *string        `yaml:"batchSheduleTimezone"`
	PodName                    *string        `yaml:"podName"`
	PodNamespace               *string        `yaml:"podNamespace"`
	WebHooks                   []WebHook      `yaml:"webhooks"`
	Snapshots                  Snapshot       `yaml:"snapshots"`
	RemoteServer               RemoteServer   `yaml:"remoteServer"`
}

func (t *Type) DeepCopy() *Type {
	copyOfType := Type{}

	typeJSON, err := json.Marshal(t)
	if err != nil {
		log.WithError(err).Fatal("error while json.Marshal")
	}

	err = json.Unmarshal(typeJSON, &copyOfType)
	if err != nil {
		log.WithError(err).Fatal("error while json.Unmarshal")
	}

	return &copyOfType
}

func Load() error {
	if len(*config.ConfigPath) == 0 {
		log.Debug("config file not set - nothing to load")

		return nil
	}

	configByte, err := os.ReadFile(*config.ConfigPath)
	if err != nil {
		return errors.Wrap(err, "can not load config file")
	}

	err = yaml.Unmarshal(configByte, &config)
	if err != nil {
		return errors.Wrap(err, "error while yaml.Unmarshal")
	}

	loadDefaults()

	return nil
}

func loadDefaults() {
	// load default project profiles
	if len(config.ProjectProfiles) == 0 {
		log.Warning("adding default project profiles")

		defaultProfile := ProjectProfile{
			Name:            "default",
			NamespacePrefix: fmt.Sprintf("%s-", Namespace),
		}

		config.ProjectProfiles = []*ProjectProfile{&defaultProfile}
	}

	if config.Links == nil {
		return
	}

	for id := range config.KubernetesEndpoints {
		if config.KubernetesEndpoints[id].Links == nil {
			config.KubernetesEndpoints[id].Links = &Links{}
		}

		if len(config.KubernetesEndpoints[id].Links.SentryURL) == 0 {
			config.KubernetesEndpoints[id].Links.SentryURL = config.Links.SentryURL
		}

		if len(config.KubernetesEndpoints[id].Links.LogsURL) == 0 {
			config.KubernetesEndpoints[id].Links.LogsURL = config.Links.LogsURL
		}

		if len(config.KubernetesEndpoints[id].Links.PhpMyAdminURL) == 0 {
			config.KubernetesEndpoints[id].Links.PhpMyAdminURL = config.Links.PhpMyAdminURL
		}

		if len(config.KubernetesEndpoints[id].Links.SlackURL) == 0 {
			config.KubernetesEndpoints[id].Links.SlackURL = config.Links.SlackURL
		}

		if len(config.KubernetesEndpoints[id].Links.MetricsURL) == 0 {
			config.KubernetesEndpoints[id].Links.MetricsURL = config.Links.MetricsURL
		}

		if len(config.KubernetesEndpoints[id].Links.TracingURL) == 0 {
			config.KubernetesEndpoints[id].Links.TracingURL = config.Links.TracingURL
		}

		if len(config.KubernetesEndpoints[id].Links.LogsPodURL) == 0 {
			config.KubernetesEndpoints[id].Links.LogsPodURL = config.Links.LogsPodURL
		}
	}
}

func CheckConfig() error {
	_, err := time.LoadLocation(*config.BatchSheduleTimezone)
	if err != nil {
		return errors.Wrap(err, "error in parsing timezone")
	}

	if len(*config.PodName) == 0 || len(*config.PodNamespace) == 0 {
		return errors.New("pod name or namespace is empty")
	}

	return nil
}

func Get() *Type {
	return &config
}

func String() string {
	out, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Sprintf("ERROR: %t", err)
	}

	return string(out)
}

func GetEnvDefault(name string, defaultValue string) string {
	r := os.Getenv(name)
	defaultValueLen := len(defaultValue)

	if defaultValueLen == 0 {
		return r
	}

	if len(r) == 0 {
		return defaultValue
	}

	return r
}

var gitVersion = "dev"

func GetVersion() string {
	return gitVersion
}

func GetProjectProfileByNamespace(namespace string) *ProjectProfile {
	for _, projectProfile := range config.ProjectProfiles {
		if strings.HasPrefix(namespace, projectProfile.NamespacePrefix) {
			return projectProfile
		}
	}

	return nil
}

func GetProjectProfileByName(name string) *ProjectProfile {
	for _, projectProfile := range config.ProjectProfiles {
		if projectProfile.Name == name {
			return projectProfile
		}
	}

	return nil
}

func GetKubernetesEndpointByName(name string) *KubernetesEndpoint {
	for _, kubernetesEndpoints := range config.KubernetesEndpoints {
		if kubernetesEndpoints.Name == name {
			return kubernetesEndpoints
		}
	}

	return nil
}
