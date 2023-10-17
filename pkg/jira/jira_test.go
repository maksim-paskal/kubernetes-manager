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
package jira_test

import (
	"context"
	"os"
	"testing"

	"github.com/maksim-paskal/kubernetes-manager/pkg/jira"
)

func TestJira(t *testing.T) {
	t.Parallel()

	if _, ok := os.LookupEnv("JIRA_URL"); !ok {
		t.Skip("Skipping test because JIRA_URL variable not set")
	}

	if _, ok := os.LookupEnv("JIRA_TOKEN"); !ok {
		t.Skip("Skipping test because JIRA_TOKEN variable not set")
	}

	if _, ok := os.LookupEnv("JIRA_ISSUE"); !ok {
		t.Skip("Skipping test because JIRA_ISSUE variable not set")
	}

	result, err := jira.GetIssueInfo(context.TODO(), os.Getenv("JIRA_ISSUE"))
	if err != nil {
		t.Fatal(err)
	}

	if result.Fields.Status.Name == "" {
		t.Fatal("result.Status is empty")
	}
}
