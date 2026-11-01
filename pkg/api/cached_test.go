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
package api

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

// TestIsPodFinished checks that evicted (Failed) pods are not counted as pods of a running environment.
func TestIsPodFinished(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		phase corev1.PodPhase
		want  bool
	}{
		{name: "running", phase: corev1.PodRunning, want: false},
		{name: "pending", phase: corev1.PodPending, want: false},
		{name: "succeeded", phase: corev1.PodSucceeded, want: true},
		{name: "failed (evicted)", phase: corev1.PodFailed, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pod := corev1.Pod{Status: corev1.PodStatus{Phase: tt.phase}}
			if got := isPodFinished(pod); got != tt.want {
				t.Fatalf("isPodFinished(%s) = %v, want %v", tt.phase, got, tt.want)
			}
		})
	}
}
