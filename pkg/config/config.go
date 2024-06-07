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
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	defaultAddr                      = ":9000"
	defaultRemoveBranchLastScaleDate = 10
	defaultBatchShedulePeriodSeconds = 30 * 60 // 30 minutes
	defaultGracefulShutdownSeconds   = 5
	defaultDelayHours                = 10

	TemporaryTokenRandLength    = 5
	TemporaryTokenDurationHours = 10

	StaledNewNamespaceDurationDays = 3

	HoursInDay     = 24
	KeyValueLength = 2

	TrueValue = "true"

	LabelScaleDownDelayShort = "scaleDownDelay"

	Namespace             = "kubernetes-manager"
	AnnotationPrefix      = Namespace + "/"
	FilterLabels          = Namespace + "=true"
	LabelType             = Namespace + "/type"
	LabelScaleDownDelay   = Namespace + "/" + LabelScaleDownDelayShort
	LabelLastScaleDate    = Namespace + "/lastScaleDate"
	LabelGitBranch        = Namespace + "/git-branch"
	LabelGitProjectID     = Namespace + "/git-project-id"
	LabelGitProjectOrigin = Namespace + "/git-project-origin"
	LabelRegistryTag      = Namespace + "/registry-tag"
	LabelSystemNamespace  = Namespace + "/system-namespace"
	TagNamespace          = Namespace + "/namespace"
	TagCluster            = Namespace + "/cluster"
	LabelNamespaceCreator = Namespace + "/user-creator"
	LabelProjectProfile   = Namespace + "/profile"
	LabelInstalledProject = Namespace + "/project"
	LabelEnvironmentName  = Namespace + "/environment-name"
	LabelUserLiked        = Namespace + "/user-liked"
	LabelGitSyncOrigin    = Namespace + "/git-sync-origin"
	LabelGitSyncBranch    = Namespace + "/git-sync-branch"
	LabelDescription      = Namespace + "/description"

	HeaderOwner = "X-Owner"
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

type Cache struct {
	Type   string
	Config interface{}
}

type KubernetesEndpoint struct {
	Disabled bool
	// set maintenance mode if cluster has some problems
	Maintenance       bool
	Name              string
	KubeConfigPath    string
	KubeConfigServer  string
	Links             *Links
	PipelineVariables map[string]string
}

type ProjectSetting struct {
	ProjectID   string
	ImagePrefix string
}

type ProjectProfileNameType string

const (
	ProjectProfileNameTypeSimple    ProjectProfileNameType = "simple"
	ProjectProfileNameTypeService   ProjectProfileNameType = "service-name"
	ProjectProfileNameTypeJiraIssue ProjectProfileNameType = "jira-issue"
)

type ProjectProfile struct {
	Name              string
	NamespaceNameType ProjectProfileNameType
	NamespacePrefix   string
	DefaultPriority   int
	SortPriorities    string // sort priority (comma separated format projectId=number)
	DefaultBranch     string // use default branch for project (comma separated format projectId=main)
	Required          string // project ids to be required (comma separated)
	Exclude           string // project ids to exclude (comma separated) or * for all
	Include           string // project ids to include (comma separated)
	PipelineVariables map[string]string
}

func (p *ProjectProfile) Validate() error {
	if p.NamespaceNameType != ProjectProfileNameTypeSimple &&
		p.NamespaceNameType != ProjectProfileNameTypeService &&
		p.NamespaceNameType != ProjectProfileNameTypeJiraIssue {
		return errors.New("invalid NamespaceNameType: " + string(p.NamespaceNameType))
	}

	if re := regexp.MustCompile(`^\d+(=|=-)\d+(,\d+(=|=-)\d+)*$`); len(p.SortPriorities) > 0 && !re.MatchString(p.SortPriorities) {
		return errors.Errorf("invalid SortPriorities, valid (%s) got (%s)", re.String(), p.SortPriorities)
	}

	if re := regexp.MustCompile(`^\d+=[0-9A-Za-z-_]+(,\d+=[0-9A-Za-z-_]+)*$`); len(p.DefaultBranch) > 0 && !re.MatchString(p.DefaultBranch) {
		return errors.Errorf("invalid DefaultBranch, valid (%s) got (%s)", re.String(), p.DefaultBranch)
	}

	if re := regexp.MustCompile(`^\d+(,\d+)*$`); len(p.Required) > 0 && !re.MatchString(p.Required) {
		return errors.Errorf("invalid Required, valid (%s) got (%s)", re.String(), p.Required)
	}

	if re := regexp.MustCompile(`^(\d+(,\d+)*|\*)$`); len(p.Exclude) > 0 && !re.MatchString(p.Exclude) {
		return errors.Errorf("invalid Exclude, valid (%s) got (%s)", re.String(), p.Exclude)
	}

	if re := regexp.MustCompile(`^\d+(,\d+)*$`); len(p.Include) > 0 && !re.MatchString(p.Include) {
		return errors.Errorf("invalid Include, valid (%s) got (%s)", re.String(), p.Include)
	}

	return nil
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
	return slices.Contains(p.GetRequired(), strconv.Itoa(projectID))
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

func (p *ProjectProfile) GetProjectSortPriority(projectID int) int {
	if len(p.SortPriorities) == 0 {
		return p.DefaultPriority
	}

	for _, sortPriority := range strings.Split(p.SortPriorities, ",") {
		sortPriorityData := strings.Split(sortPriority, "=")
		if len(sortPriorityData) != KeyValueLength {
			log.Errorf("invalid sortPriority format %s", sortPriority)

			continue
		}

		if sortPriorityData[0] == strconv.Itoa(projectID) {
			sortPriorityInt, err := strconv.Atoi(sortPriorityData[1])
			if err != nil {
				log.WithError(err).Errorf("error while converting sortPriority %s to int", sortPriorityData[1])

				continue
			}

			return sortPriorityInt
		}
	}

	return p.DefaultPriority
}

type NamespaceMeta struct {
	Pattern     string
	Labels      map[string]string
	Annotations map[string]string
}

func (n *NamespaceMeta) GetTemplatedValue(ctx context.Context) *NamespaceMeta {
	newMeta := *n

	for key, value := range n.Labels {
		if templatedValue, err := utils.GetTemplatedResult(ctx, value, config); err == nil {
			newMeta.Labels[key] = string(templatedValue)
		} else {
			log.WithError(err).Errorf("error while templating label %s", key)
		}
	}

	for key, value := range n.Annotations {
		if templatedValue, err := utils.GetTemplatedResult(ctx, value, config); err == nil {
			newMeta.Annotations[key] = string(templatedValue)
		} else {
			log.WithError(err).Errorf("error while templating annotation %s", key)
		}
	}

	return &newMeta
}

type WebHook struct {
	Provider string
	Config   interface{}
	IDs      []string
	Events   []types.Event
}

type Snapshot struct {
	ProjectID string
	Ref       string
}

type RemoteServer struct {
	HetznerToken string
	Links        []*OtherLink
}

type AutotestCustomActionEnvType string

const (
	AutotestCustomActionEnvList AutotestCustomActionEnvType = "list"
	AutotestCustomActionEnvText AutotestCustomActionEnvType = "text"
)

type AutotestCustomActionEnv struct {
	Name        string
	Default     string
	Description string
	Type        AutotestCustomActionEnvType
}

type AutotestCustomAction struct {
	ProjectID int
	Tests     []string
	Env       []*AutotestCustomActionEnv
}

func (d *AutotestCustomAction) DeepCopy() *AutotestCustomAction {
	copyOfCustomAction := AutotestCustomAction{}

	if d == nil {
		return &copyOfCustomAction
	}

	typeJSON, err := json.Marshal(d)
	if err != nil {
		log.WithError(err).Fatal("error while json.Marshal")
	}

	err = json.Unmarshal(typeJSON, &copyOfCustomAction)
	if err != nil {
		log.WithError(err).Fatal("error while json.Unmarshal")
	}

	return &copyOfCustomAction
}

type AutotestAction struct {
	Name    string
	Test    string
	Release string
	Ref     string
}
type Autotest struct {
	Pattern           string
	ProjectID         int
	ReportURL         string
	Actions           []*AutotestAction
	CustomAction      *AutotestCustomAction
	FilterByNamespace bool
}

func (a *Autotest) GetActionByTest(test string) *AutotestAction {
	for _, action := range a.Actions {
		if action.Test == test {
			return action
		}
	}

	return nil
}

func (t *Type) GetAutotestByID(id string) *Autotest {
	for _, a := range t.Autotests {
		if regexp.MustCompile(a.Pattern).Match([]byte(id)) {
			return a
		}
	}

	return nil
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

	WebListen: flag.String("web.listen", defaultAddr, ""),
	FrontDist: flag.String("front.dist", "front/dist", ""),

	PodName:      flag.String("pod.name", os.Getenv("POD_NAME"), ""),
	PodNamespace: flag.String("pod.namespace", os.Getenv("POD_NAMESPACE"), ""),

	BatchShedulePeriodSeconds: flag.Int("batch.periodSeconds", defaultBatchShedulePeriodSeconds, "batch shedule period"),

	GitlabToken:     flag.String("gitlab.token", os.Getenv("GITLAB_TOKEN"), ""),
	GitlabTokenUser: flag.String("gitlab.token.user", os.Getenv("GITLAB_TOKEN_USER"), "username of token user (need to filter pipelines)"),
	GitlabURL:       flag.String("gitlab.url", os.Getenv("GITLAB_URL"), ""),

	IngressHostDefaultProtocol: flag.String("ingress.show-protocol", "https", ""),

	RemoveBranchLastScaleDate: flag.Int("batch.removeBranchLastScaleDate", defaultRemoveBranchLastScaleDate, ""),

	ExternalServicesTopic: flag.String("externalServicesTopic", GetEnvDefault("EXTERNAL_SERVICES_TOPIC", "kubernetes-manager"), ""),
	BatchEnabled:          flag.Bool("batch.enabled", true, "enable batch operations"),

	GracefulShutdownSeconds: flag.Int("gracefulShutdownSeconds", defaultGracefulShutdownSeconds, "graceful shutdown timeout"),

	DelayHours: flag.Int("delayHours", defaultDelayHours, "default delay hours"),
	Cache: &Cache{
		Type: "noop",
	},
}

type Type struct {
	GracefulShutdownSeconds *int

	ConfigPath                 *string
	LogLevel                   *string
	Links                      *Links
	BatchEnabled               *bool
	NamespaceMeta              []*NamespaceMeta
	DebugTemplates             []*Template
	ProjectProfiles            []*ProjectProfile
	ProjectSettings            []*ProjectSetting
	KubernetesEndpoints        []*KubernetesEndpoint
	WebListen                  *string
	FrontDist                  *string
	RemoveBranchDaysInactive   *int
	GitlabToken                *string
	GitlabTokenUser            *string
	GitlabURL                  *string
	IngressHostDefaultProtocol *string
	RemoveBranchLastScaleDate  *int
	ExternalServicesTopic      *string
	BatchShedulePeriodSeconds  *int
	PodName                    *string
	PodNamespace               *string
	WebHooks                   []WebHook
	Snapshots                  Snapshot
	RemoteServer               RemoteServer
	Autotests                  []*Autotest
	DelayHours                 *int
	WikiPages                  []*WikiPage
	Cache                      *Cache
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

func (t *Type) GetDefaultDelay() string {
	delayHours, err := time.ParseDuration(fmt.Sprintf("%dh", *t.DelayHours))
	if err != nil {
		return utils.TimeToString(time.Now())
	}

	return utils.TimeToString(time.Now().Add(delayHours))
}

func (t *Type) GetProjectSetting(projectID string) *ProjectSetting {
	for _, projectSetting := range t.ProjectSettings {
		if projectSetting.ProjectID == projectID {
			return projectSetting
		}
	}

	return nil
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
			NamespacePrefix: Namespace + "-",
		}

		config.ProjectProfiles = []*ProjectProfile{&defaultProfile}
	}

	for id := range config.ProjectProfiles {
		if len(config.ProjectProfiles[id].NamespaceNameType) == 0 {
			config.ProjectProfiles[id].NamespaceNameType = ProjectProfileNameTypeJiraIssue
		}
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

		if len(config.KubernetesEndpoints[id].Links.JiraURL) == 0 {
			config.KubernetesEndpoints[id].Links.JiraURL = config.Links.JiraURL
		}
	}
}

func CheckConfig() error {
	if len(*config.PodName) == 0 || len(*config.PodNamespace) == 0 {
		return errors.New("pod name or namespace is empty")
	}

	for _, profile := range config.ProjectProfiles {
		if err := profile.Validate(); err != nil {
			return errors.Wrap(err, "error while validating profile: "+profile.Name)
		}
	}

	return nil
}

func Get() *Type {
	return &config
}

func (t *Type) String() string {
	out, err := json.Marshal(t)
	if err != nil {
		return fmt.Sprintf("ERROR: %t", err)
	}

	return string(out)
}

func (t *Type) GetBatchShedulePeriod() time.Duration {
	return time.Duration(*t.BatchShedulePeriodSeconds) * time.Second
}

func (t *Type) GetGracefulShutdown() time.Duration {
	return time.Duration(*t.GracefulShutdownSeconds) * time.Second
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

func (t *Type) GetKubernetesEndpoints() []*KubernetesEndpoint {
	result := make([]*KubernetesEndpoint, 0)

	for _, kubernetesEndpoints := range t.KubernetesEndpoints {
		if kubernetesEndpoints.Disabled {
			continue
		}

		result = append(result, kubernetesEndpoints)
	}

	return result
}

func (t *Type) GetKubernetesEndpointByName(name string) *KubernetesEndpoint {
	for _, kubernetesEndpoints := range t.KubernetesEndpoints {
		if kubernetesEndpoints.Name == name {
			return kubernetesEndpoints
		}
	}

	return nil
}

func GetNamespaceMeta(ctx context.Context, namespace string) *NamespaceMeta {
	for _, namespaceMeta := range config.DeepCopy().NamespaceMeta {
		if len(namespaceMeta.Pattern) == 0 {
			return namespaceMeta.GetTemplatedValue(ctx)
		} else if regexp.MustCompile(namespaceMeta.Pattern).Match([]byte(namespace)) {
			return namespaceMeta.GetTemplatedValue(ctx)
		}
	}

	// if not found, return empty meta
	return &NamespaceMeta{
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}
}
