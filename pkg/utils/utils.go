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
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	convertStringToInt64Base    = 10
	convertStringToInt64BitSize = 32
	timeFormat                  = time.RFC3339
	hoursInDay                  = 24
	keyValueLength              = 2
)

func TimeToString(t time.Time) string {
	return t.Format(timeFormat)
}

func StringToTime(value string) (time.Time, error) {
	return time.Parse(timeFormat, value)
}

// returns hours between current time and input time.
func DiffToNowHours(t time.Time) int {
	now := time.Now()

	return int(now.Sub(t).Hours())
}

// returns days between current time and input time.
func DiffToNowDays(t time.Time) int {
	return DiffToNowHours(t) / hoursInDay
}

func ConvertStringToInt64(value string) (int64, error) {
	return strconv.ParseInt(value, convertStringToInt64Base, convertStringToInt64BitSize)
}

func RandomString(l int) string {
	buff := make([]byte, int(math.Ceil(float64(l)/keyValueLength)))
	_, _ = rand.Read(buff)
	str := hex.EncodeToString(buff)

	return str[:l]
}

func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}

	return false
}

type JaegerLogs struct{}

func (l JaegerLogs) Error(msg string) {
	log.Errorf(msg)
}

func (l JaegerLogs) Infof(msg string, args ...interface{}) {
	log.Debugf(msg, args...)
}
