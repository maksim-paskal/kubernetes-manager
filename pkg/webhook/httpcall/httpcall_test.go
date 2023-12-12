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
package httpcall_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/maksim-paskal/kubernetes-manager/pkg/webhook/httpcall"
	"github.com/pkg/errors"
)

func testHandler(t *testing.T) *mux.Router {
	t.Helper()

	mux := mux.NewRouter()

	test := func(r *http.Request) error {
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			return errors.Errorf("bad content type: %s", contentType)
		}

		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return errors.Wrap(err, "io.ReadAll")
		}

		if !strings.Contains(string(body), "testEvent") {
			return errors.Errorf("bad body: %s", string(body))
		}

		return nil
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := test(r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})

	return mux
}

func TestNotify(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(testHandler(t))

	provider := httpcall.Provider{}

	conditions := config.WebHook{
		Provider: "notify",
		Config: httpcall.ProviderConfig{
			URL: ts.URL,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `
{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "{{ .Message.Event }}"
			}
		}
	]
}`,
		},
	}
	message := types.WebhookMessage{
		Event:     "testEvent",
		Name:      "testName",
		Cluster:   "testCluster",
		Namespace: "testNamespace",
	}

	if err := provider.Init(conditions, message); err != nil {
		t.Fatal(err)
	}

	if err := provider.Process(context.TODO()); err != nil {
		t.Fatal(err)
	}
}
