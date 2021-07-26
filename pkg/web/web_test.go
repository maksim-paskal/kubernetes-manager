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
package web_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maksim-paskal/kubernetes-manager/pkg/web"
)

var (
	client = &http.Client{}
	ts     = httptest.NewServer(web.GetHandler())
	ctx    = context.Background()
)

func TestVersion(t *testing.T) {
	t.Parallel()

	url := fmt.Sprintf("%s/api/version", ts.URL)
	t.Log(url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Log(string(body))

	version := web.APIVersion{}

	err = json.Unmarshal(body, &version)
	if err != nil {
		t.Fatal(err)
	}

	if version.Version != "dev" {
		t.Fatal("no version")
	}
}
