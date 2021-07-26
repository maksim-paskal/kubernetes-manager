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

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	defaultPort                      = 9000
	defaultRemoveBranchLastScaleDate = 10
	defaultNotDeleteDays             = 10
	defaultRemoveBranchDaysInactive  = 20

	HoursInDay            = 24
	KeyValueLength        = 2
	LabelLastScaleDate    = "kubernetes-manager/lastScaleDate"
	LabelGitBranch        = "kubernetes-manager/git-branch"
	LabelGitProjectID     = "kubernetes-manager/git-project-id"
	LabelGitProjectOrigin = "kubernetes-manager/git-project-origin"
	LabelRegistryTag      = "kubernetes-manager/registry-tag"
)

//nolint:gochecknoglobals
var config = Type{
	ConfigPath: flag.String("config", "config.yaml", "config"),
	LogLevel:   flag.String("log.level", "INFO", "logging level"),

	KubeConfigServer: flag.String("kubeconfig.server", os.Getenv("KUBERNETES_ENDPOINT"), ""),
	KubeConfigPath:   flag.String("kubeconfig.path", "", "path to kubeconfig"),

	Mode:                     flag.String("mode", "web", "run mode"),
	Port:                     flag.Int("server.port", defaultPort, ""),
	FrontDist:                flag.String("front.dist", "front/dist", ""),
	RemoveBranchDaysInactive: flag.Int("batch.removeBranchDaysInactive", defaultRemoveBranchDaysInactive, ""),

	GitlabToken: flag.String("gitlab.token", os.Getenv("GITLAB_TOKEN"), ""),
	GitlabURL:   flag.String("gitlab.url", os.Getenv("GITLAB_URL"), ""),

	IngressHostDefaultProtocol: flag.String("ingress.show-protocol", "https", ""),
	IngressFilter:              flag.String("ingress.filter", "kubernetes-manager=true", ""),
	IngressNoFiltration:        flag.Bool("ingress.no-filtration", false, ""),

	RemoveBranchLastScaleDate: flag.Int("batch.removeBranchLastScaleDate", defaultRemoveBranchLastScaleDate, ""),

	RegistryDirectory: flag.String("registry.directory", os.Getenv("REGISTRY_DIRECTORY"), ""),
	RegistryURL:       flag.String("registry.url", os.Getenv("REGISTRY_URL"), ""),
	RegistryUser:      flag.String("registry.user", os.Getenv("REGISTRY_USER"), ""),
	RegistryPassword:  flag.String("registry.password", os.Getenv("REGISTRY_PASSWORD"), ""),

	ReleasePatern:        flag.String("release.pattern", `release-(\d{4}\d{2}\d{2}).*`, ""),
	ReleaseNotDeleteDays: flag.Int("release.notDeleteDays", defaultNotDeleteDays, ""),

	MakeAPICallServer: flag.String("makeAPICall.server", "127.0.0.1", ""),

	SystemNamespaces: flag.String("system.namespaces", getEnvDefault("SYSTEM_NAMESPACES", "^kube-system$"), ""),
	SystemGitTags:    flag.String("system.gitTags", getEnvDefault("SYSTEM_GIT_TAGS", "^master$,^release-.*"), ""),
}

type Type struct {
	ConfigPath                 *string `yaml:"configPath"`
	LogLevel                   *string `yaml:"logLevel"`
	KubeConfigServer           *string `yaml:"kubeConfigServer"`
	KubeConfigPath             *string `yaml:"kubeConfigPath"`
	Mode                       *string `yaml:"mode"`
	Port                       *int    `yaml:"port"`
	FrontDist                  *string `yaml:"frontDist"`
	RemoveBranchDaysInactive   *int    `yaml:"removeBranchDaysInactive"`
	GitlabToken                *string `yaml:"gitlabToken"`
	GitlabURL                  *string `yaml:"gitlabUrl"`
	IngressFilter              *string `yaml:"ingressFilter"`
	IngressNoFiltration        *bool   `yaml:"ingressNoFiltration"`
	IngressHostDefaultProtocol *string `yaml:"ingressHostDefaultProtocol"`
	RemoveBranchLastScaleDate  *int    `yaml:"removeBranchLastScaleDate"`
	RegistryDirectory          *string `yaml:"registryDirectory"`
	RegistryURL                *string `yaml:"registryUrl"`
	RegistryUser               *string `yaml:"registryUser"`
	RegistryPassword           *string `yaml:"registryPassword"`
	ReleasePatern              *string `yaml:"releasePatern"`
	ReleaseNotDeleteDays       *int    `yaml:"releaseNotDeleteDays"`
	MakeAPICallServer          *string `yaml:"makeApiCallServer"`
	SystemNamespaces           *string `yaml:"systemNamespaces"`
	SystemGitTags              *string `yaml:"systemGitTags"`
}

func Load() error {
	configByte, err := ioutil.ReadFile(*config.ConfigPath)
	if err != nil {
		log.Debug(err)

		return nil
	}

	err = yaml.Unmarshal(configByte, &config)
	if err != nil {
		return err
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

func getEnvDefault(name string, defaultValue string) string {
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
