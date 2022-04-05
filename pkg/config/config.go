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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	defaultPort                      = 9000
	defaultRemoveBranchLastScaleDate = 10
	defaultRemoveBranchDaysInactive  = 20
	defaultBatchShedulePeriod        = 30 * time.Minute
	defaultBatchTimezone             = "UTC"

	ScaleDownHourMinPeriod = 19
	ScaleDownHourMaxPeriod = 5

	HoursInDay            = 24
	KeyValueLength        = 2
	LabelScaleDownDelay   = "kubernetes-manager/scaleDownDelay"
	LabelLastScaleDate    = "kubernetes-manager/lastScaleDate"
	LabelGitBranch        = "kubernetes-manager/git-branch"
	LabelGitProjectID     = "kubernetes-manager/git-project-id"
	LabelGitProjectOrigin = "kubernetes-manager/git-project-origin"
	LabelRegistryTag      = "kubernetes-manager/registry-tag"
)

type Links struct {
	SentryURL     string `yaml:"sentryUrl"`
	SlackURL      string `yaml:"slackUrl"`
	LogsURL       string `yaml:"logsUrl"`
	PhpMyAdminURL string `yaml:"phpMyAdminUrl"`
	MetricsURL    string `yaml:"metricsUrl"`
	TracingURL    string `yaml:"tracingUrl"`
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
	Links            Links
}

//nolint:gochecknoglobals
var config = Type{
	ConfigPath: flag.String("config", os.Getenv("CONFIG"), "config"),
	LogLevel:   flag.String("log.level", "INFO", "logging level"),

	KubernetesEndpoints: []KubernetesEndpoint{{
		Name:             "default",
		KubeConfigServer: GetEnvDefault("DEFAULT_CONFIG_SERVER", "https://cluster-public"),
		KubeConfigPath:   "",
	}},

	Port:                     flag.Int("server.port", defaultPort, ""),
	FrontDist:                flag.String("front.dist", "front/dist", ""),
	RemoveBranchDaysInactive: flag.Int("batch.removeBranchDaysInactive", defaultRemoveBranchDaysInactive, ""),
	ExecuteBatch:             flag.Bool("executeBatch", false, "execute Batch"),
	BatchShedulePeriod:       flag.Duration("batch.period", defaultBatchShedulePeriod, "batch shedule period"),
	BatchSheduleTimezone:     flag.String("batch.timeZone", defaultBatchTimezone, "batch shedule timezone"),

	GitlabToken: flag.String("gitlab.token", os.Getenv("GITLAB_TOKEN"), ""),
	GitlabURL:   flag.String("gitlab.url", os.Getenv("GITLAB_URL"), ""),

	IngressHostDefaultProtocol: flag.String("ingress.show-protocol", "https", ""),
	IngressFilter:              flag.String("ingress.filter", "kubernetes-manager=true", ""),
	IngressNoFiltration:        flag.Bool("ingress.no-filtration", false, ""),

	RemoveBranchLastScaleDate: flag.Int("batch.removeBranchLastScaleDate", defaultRemoveBranchLastScaleDate, ""),

	SystemNamespaces: flag.String("system.namespaces", GetEnvDefault("SYSTEM_NAMESPACES", "^kube-system$"), ""),
	SystemGitTags:    flag.String("system.gitTags", GetEnvDefault("SYSTEM_GIT_TAGS", "^master$|^release-.*"), ""),

	ExternalServicesTopic: flag.String("externalServicesTopic", GetEnvDefault("EXTERNAL_SERVICES_TOPIC", "kubernetes-manager"), ""), //nolint:lll
}

type Type struct {
	ConfigPath                 *string `yaml:"configPath"`
	LogLevel                   *string `yaml:"logLevel"`
	Links                      Links   `yaml:"links"`
	DebugTemplates             []Template
	ExternalServicesTemplates  []Template
	KubernetesEndpoints        []KubernetesEndpoint
	Port                       *int           `yaml:"port"`
	FrontDist                  *string        `yaml:"frontDist"`
	RemoveBranchDaysInactive   *int           `yaml:"removeBranchDaysInactive"`
	GitlabToken                *string        `yaml:"gitlabToken"`
	GitlabURL                  *string        `yaml:"gitlabUrl"`
	IngressFilter              *string        `yaml:"ingressFilter"`
	IngressNoFiltration        *bool          `yaml:"ingressNoFiltration"`
	IngressHostDefaultProtocol *string        `yaml:"ingressHostDefaultProtocol"`
	RemoveBranchLastScaleDate  *int           `yaml:"removeBranchLastScaleDate"`
	SystemNamespaces           *string        `yaml:"systemNamespaces"`
	SystemGitTags              *string        `yaml:"systemGitTags"`
	ExternalServicesTopic      *string        `yaml:"externalServicesTopic"`
	BatchShedulePeriod         *time.Duration `yaml:"batchShedulePeriod"`
	BatchSheduleTimezone       *string        `yaml:"batchSheduleTimezone"`
	ExecuteBatch               *bool          `yaml:"executeBatch"`
}

func Load() error {
	if len(*config.ConfigPath) == 0 {
		log.Debug("config file not set - nothing to load")

		return nil
	}

	configByte, err := ioutil.ReadFile(*config.ConfigPath)
	if err != nil {
		return errors.Wrap(err, "can not load config file")
	}

	err = yaml.Unmarshal(configByte, &config)
	if err != nil {
		return errors.Wrap(err, "error while yaml.Unmarshal")
	}

	loadDefaults(config)

	return nil
}

func loadDefaults(config Type) {
	for id := range config.KubernetesEndpoints {
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
	}
}

func CheckConfig() error {
	_, err := time.LoadLocation(*config.BatchSheduleTimezone)
	if err != nil {
		return errors.Wrap(err, "error in parsing timezone")
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
