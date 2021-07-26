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
package utils

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	log "github.com/sirupsen/logrus"
)

const (
	convertStringToInt64Base    = 10
	convertStringToInt64BitSize = 32
)

func IsSystemBranch(gitBranch string) bool {
	for _, gitBranchRegexp := range strings.Split(*config.Get().SystemGitTags, ",") {
		re := regexp.MustCompile(gitBranchRegexp)
		if re.MatchString(strings.ToLower(gitBranch)) {
			return true
		}
	}

	return false
}

func IsSystemNamespace(namespace string) bool {
	for _, namespaceRegexp := range strings.Split(*config.Get().SystemNamespaces, ",") {
		re := regexp.MustCompile(namespaceRegexp)
		if re.MatchString(strings.ToLower(namespace)) {
			return true
		}
	}

	return false
}

// returns hours between current time and input time.
func DiffToNow(t time.Time) int {
	t1 := time.Now()

	return int(t1.Sub(t).Hours() / config.HoursInDay)
}

// returns defaultValue if env with name not found.
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

func ConvertStringToInt64(value string) (int64, error) {
	return strconv.ParseInt(value, convertStringToInt64Base, convertStringToInt64BitSize)
}

type JaegerLogs struct{}

func (l JaegerLogs) Error(msg string) {
	log.Errorf(msg)
}

func (l JaegerLogs) Infof(msg string, args ...interface{}) {
	log.Debugf(msg, args...)
}
