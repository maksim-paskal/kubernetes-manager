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
package main

import (
	"os"
	"regexp"
	"strings"
	"time"
)

func isSystemBranch(gitBranch string) bool {
	for _, gitBranchRegexp := range strings.Split(*appConfig.systemGitTags, ",") {
		re := regexp.MustCompile(gitBranchRegexp)
		if re.MatchString(strings.ToLower(gitBranch)) {
			return true
		}
	}
	return false
}

func isSystemNamespace(namespace string) bool {
	for _, namespaceRegexp := range strings.Split(*appConfig.systemNamespaces, ",") {
		re := regexp.MustCompile(namespaceRegexp)
		if re.MatchString(strings.ToLower(namespace)) {
			return true
		}
	}
	return false
}

func diffToNow(t time.Time) int {
	t1 := time.Now()
	return int(t1.Sub(t).Hours() / 24)
}

func getEnvDefault(name string, defaultValue string) string {
	r := os.Getenv(name)
	if len(defaultValue) == 0 {
		return r
	}
	if len(r) == 0 {
		return defaultValue
	}
	return r
}
