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
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestConcurrentRequests runs slow requests in parallel on one client, as the parallel --wait does.
// The requests last longer than the second before the spinner starts, so the spinner is used by all of them.
func TestConcurrentRequests(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(1200 * time.Millisecond)

		_ = json.NewEncoder(w).Encode(map[string]any{"Result": map[string]any{"PodsTotal": 1, "PodsReady": 1}})
	}))
	t.Cleanup(server.Close)

	cli := NewClient().WithBaseURL(server.URL).WithBasicAuth("login", "password")

	var wg sync.WaitGroup

	for _, id := range []string{"c:one", "c:two", "c:three"} {
		wg.Go(func() {
			if _, err := cli.GetPodsInfo(context.Background(), id); err != nil {
				t.Error(err)
			}
		})
	}

	wg.Wait()
}
