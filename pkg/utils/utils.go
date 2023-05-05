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
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	convertStringToInt64Base    = 10
	convertStringToInt64BitSize = 32
	timeFormat                  = time.RFC3339
	hoursInDay                  = 24
	keyValueLength              = 2
)

func TimeToUnix(t time.Time) string {
	return strconv.Itoa(int(t.Unix()))
}

func UnixToTime(value string) (time.Time, error) {
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Now(), err
	}

	return time.Unix(i, 0), nil
}

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
	if len(list) == 0 {
		return false
	}

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

func GetTemplatedResult(text string, obj interface{}) ([]byte, error) {
	t, err := template.New("getTemplatedString").Funcs(sprig.FuncMap()).Parse(text)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing template")
	}

	buf := &bytes.Buffer{}

	err = t.Execute(buf, &obj)
	if err != nil {
		return nil, errors.Wrap(err, "error executing template")
	}

	return buf.Bytes(), nil
}

type HumanizeDurationType string

const (
	HumanizeDurationShort HumanizeDurationType = "short"
	HumanizeDurationFull  HumanizeDurationType = "full"
)

func HumanizeDuration(showType HumanizeDurationType, duration time.Duration) string {
	const (
		secondsInMinute = 60
		minutesInHour   = 60
		hoursInDay      = 24
	)

	if duration.Seconds() < secondsInMinute {
		return fmt.Sprintf("%d seconds", int64(duration.Seconds()))
	}

	if duration.Minutes() < minutesInHour {
		remainingSeconds := math.Mod(duration.Seconds(), minutesInHour)

		if showType == HumanizeDurationShort {
			return fmt.Sprintf("%d minutes", int64(duration.Minutes()))
		}

		return fmt.Sprintf("%d minutes %d seconds", int64(duration.Minutes()), int64(remainingSeconds))
	}

	if duration.Hours() < hoursInDay {
		remainingMinutes := math.Mod(duration.Minutes(), minutesInHour)
		remainingSeconds := math.Mod(duration.Seconds(), secondsInMinute)

		if showType == HumanizeDurationShort {
			return fmt.Sprintf("%d hours", int64(duration.Hours()))
		}

		return fmt.Sprintf("%d hours %d minutes %d seconds",
			int64(duration.Hours()), int64(remainingMinutes), int64(remainingSeconds))
	}

	remainingHours := math.Mod(duration.Hours(), hoursInDay)
	remainingMinutes := math.Mod(duration.Minutes(), minutesInHour)
	remainingSeconds := math.Mod(duration.Seconds(), secondsInMinute)

	if showType == HumanizeDurationShort {
		return fmt.Sprintf("%d days", int64(duration.Hours()/hoursInDay))
	}

	return fmt.Sprintf("%d days %d hours %d minutes %d seconds",
		int64(duration.Hours()/hoursInDay), int64(remainingHours),
		int64(remainingMinutes), int64(remainingSeconds))
}
