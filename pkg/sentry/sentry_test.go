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
package sentry_test

import (
	"context"
	"os"
	"testing"

	"github.com/maksim-paskal/kubernetes-manager/pkg/sentry"
	log "github.com/sirupsen/logrus"
)

func TestClientRighTrim(t *testing.T) {
	t.Parallel()

	test := make(map[string]string)

	test["http://test1.com/"] = "http://test1.com"
	test["http://test2.com"] = "http://test2.com"
	test["http://test3.com//"] = "http://test3.com"

	for k, v := range test {
		if sentry.NewClient(k).Endpoint != v {
			t.Fatal("test failed", k, v)
		}
	}
}

func TestClient(t *testing.T) {
	t.Parallel()

	log.SetLevel(log.DebugLevel)

	sentryClient := sentry.NewClient(os.Getenv("TEST_SENTRY_ENDPOINT"))

	sentryClient.Token = os.Getenv("TEST_SENTRY_TOKEN")
	sentryClient.Organization = os.Getenv("TEST_SENTRY_ORGANIZATION")

	if len(sentryClient.Endpoint) == 0 {
		t.Skip("no sentry endpoint")
	}

	issues, err := sentryClient.GetIssues(context.TODO(),
		os.Getenv("TEST_SENTRY_QUERY"),
		"7d",
		"25",
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, issue := range issues {
		log.Debugf("%+v", issue.LastSeenTime())
	}

	t.Fatal("test failed", len(issues))
}
