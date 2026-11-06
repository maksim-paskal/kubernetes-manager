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
		{value: "0:main", wantErr: true},
		{value: "-5:main", wantErr: true},
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

func TestCreateEnvironmentWithoutClusterAndName(t *testing.T) {
	t.Parallel()

	var body map[string]string

	var decodeErr error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		decodeErr = json.Unmarshal(raw, &body)

		_, _ = w.Write([]byte(`{"Result":"cluster:namespace"}`))
	}))
	defer server.Close()

	cli := NewClient().WithBaseURL(server.URL).WithBasicAuth("login", "password")

	// cluster and name are optional, the server defaults the cluster and skips the name when they are empty
	id, err := cli.CreateEnvironment(context.Background(), CreateEnvironmentInput{
		Profile:  "my-profile",
		Services: []ServiceRef{{ProjectID: 12, Ref: "main"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if decodeErr != nil {
		t.Fatal(decodeErr)
	}

	if got, want := id, "cluster:namespace"; got != want {
		t.Fatalf("id %q, want %q", got, want)
	}

	// the keys must be sent, otherwise the empty values below are read from a nil map
	for _, key := range []string{"Cluster", "Name"} {
		got, ok := body[key]
		if !ok {
			t.Fatalf("%s not sent", key)
		}

		if got != "" {
			t.Fatalf("%s sent %q, want empty", key, got)
		}
	}
}

func TestUserAgent(t *testing.T) {
	t.Parallel()

	var userAgent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent = r.Header.Get("User-Agent")

		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	cli := NewClient().WithBaseURL(server.URL).WithBasicAuth("login", "password").WithUserAgent("kubernetes-manager-cli/abc123 linux")

	if err := cli.DeleteEnvironment(context.Background(), "cluster:namespace"); err != nil {
		t.Fatal(err)
	}

	if got, want := userAgent, "kubernetes-manager-cli/abc123 linux"; got != want {
		t.Fatalf("user agent sent %q, want %q", got, want)
	}
}
