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
	"crypto/rand"
	"encoding/hex"
	"math"
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
	timeFormat                  = time.RFC3339
)

func IsSystemBranch(gitBranch string) bool {
	re := regexp.MustCompile(*config.Get().SystemGitTags)

	return re.MatchString(strings.ToLower(gitBranch))
}

func IsSystemNamespace(namespace string) bool {
	re := regexp.MustCompile(*config.Get().SystemNamespaces)

	return re.MatchString(strings.ToLower(namespace))
}

func TimeToString(t time.Time) string {
	return t.Format(timeFormat)
}

func StringToTime(value string) (time.Time, error) {
	return time.Parse(timeFormat, value)
}

// returns hours between current time and input time.
func DiffToNow(t time.Time) int {
	t1 := time.Now()

	return int(t1.Sub(t).Hours() / config.HoursInDay)
}

func ConvertStringToInt64(value string) (int64, error) {
	return strconv.ParseInt(value, convertStringToInt64Base, convertStringToInt64BitSize)
}

func RandomString(l int) string {
	buff := make([]byte, int(math.Ceil(float64(l)/config.KeyValueLength)))
	_, _ = rand.Read(buff)
	str := hex.EncodeToString(buff)

	return str[:l]
}

type JaegerLogs struct{}

func (l JaegerLogs) Error(msg string) {
	log.Errorf(msg)
}

func (l JaegerLogs) Infof(msg string, args ...interface{}) {
	log.Debugf(msg, args...)
}
