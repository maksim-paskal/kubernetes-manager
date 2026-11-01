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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseServiceRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		want    ServiceRef
		wantErr bool
	}{
		{value: "12:main", want: ServiceRef{ProjectID: 12, Ref: "main"}},
		{value: "15:feature/x-1", want: ServiceRef{ProjectID: 15, Ref: "feature/x-1"}},
		{value: "12", wantErr: true},
		{value: "12:", wantErr: true},
		{value: "app:main", wantErr: true},
		{value: "12:feature:x", wantErr: true},
		{value: "12:a;b", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()

			got, err := ParseServiceRef(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFormatServiceRefs(t *testing.T) {
	t.Parallel()

	if got := formatServiceRefs(nil); got != "" {
		t.Fatalf("got %q, want empty", got)
	}

	got := formatServiceRefs([]ServiceRef{{ProjectID: 12, Ref: "main"}, {ProjectID: 15, Ref: "dev"}})
	if want := "12:main;15:dev"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestScaleDownDelaySendsDuration(t *testing.T) {
	t.Parallel()

	var body map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)

		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	cli := NewClient().WithBaseURL(server.URL).WithBasicAuth("login", "password")

	if err := cli.ScaleDownDelayEnvironment(context.Background(), "cluster:namespace", 90*time.Minute); err != nil {
		t.Fatal(err)
	}

	if got, want := body["Delay"], "1h30m0s"; got != want {
		t.Fatalf("delay sent %q, want %q", got, want)
	}
}
