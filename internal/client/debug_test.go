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
	"net/http"
	"testing"
)

func TestDebugBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "json is indented", body: `{"Result":{"PodsTotal":97}}` + "\n", want: "response body\n{\n  \"Result\": {\n    \"PodsTotal\": 97\n  }\n}"},
		{name: "text is kept", body: "some error\n", want: "response body\nsome error"},
		{name: "empty", body: "", want: "response body\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := debugBody("response body", []byte(tt.body)); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRedactHeaders(t *testing.T) {
	t.Parallel()

	headers := http.Header{
		"Authorization": {"Basic secret"},
		"Set-Cookie":    {"session=secret"},
		"Cache-Control": {"max-age=10"},
	}

	got := redactHeaders(headers)

	if got.Get("Authorization") != "REDACTED" || got.Get("Set-Cookie") != "REDACTED" || got.Get("Cache-Control") != "max-age=10" {
		t.Fatalf("headers = %v", got)
	}

	// the original is not changed, it is still used by the request
	if headers.Get("Authorization") != "Basic secret" {
		t.Fatal("redactHeaders changed the original headers")
	}

	if got := redactHeaders(nil); len(got) != 0 {
		t.Fatalf("nil headers = %v", got)
	}
}
