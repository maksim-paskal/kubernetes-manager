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
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodsInfo struct {
	PodsTotal       int64
	PodsReady       int64
	PodsFailed      int64
	PodsFailedName  []string
	cpuRequests     float64
	memoryRequests  float64
	storageRequests float64

	CPURequests     string
	MemoryRequests  string
	StorageRequests string
}

func (e *Environment) GetPodsInfo(ctx context.Context) (*PodsInfo, error) {
	pods, err := e.clientset.CoreV1().Pods(e.Namespace).List(ctx, metav1.ListOptions{
		FieldSelector: "status.phase!=Succeeded",
	})
	if err != nil {
		return nil, errors.Wrap(err, "error list pods")
	}

	result := PodsInfo{}

	result.PodsFailedName = make([]string, 0)

	for _, pod := range pods.Items {
		result.PodsTotal++

		isPodReady := false

		if pod.Status.Phase == corev1.PodRunning {
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
					isPodReady = true
				}
			}
		}

		podReason := string(pod.Status.Phase)

		for _, container := range pod.Spec.Containers {
			result.cpuRequests += container.Resources.Requests.Cpu().AsApproximateFloat64()
			result.memoryRequests += container.Resources.Requests.Memory().AsApproximateFloat64()
		}

		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				if reason := containerStatus.State.Waiting.Reason; len(reason) > 0 {
					podReason = reason
				}
			}
		}

		if isPodReady {
			result.PodsReady++
		} else {
			result.PodsFailed++

			if podReason == string(corev1.PodRunning) {
				podReason = "NotReady"
			}

			podName := fmt.Sprintf("%s (%s)", pod.Name, podReason)

			result.PodsFailedName = append(result.PodsFailedName, podName)
		}
	}

	pvc, err := e.clientset.CoreV1().PersistentVolumeClaims(e.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error list pvc")
	}

	for _, pvc := range pvc.Items {
		result.storageRequests += pvc.Status.Capacity.Storage().AsApproximateFloat64()
	}

	result.CPURequests = fmt.Sprintf("%.2f", result.cpuRequests)
	result.MemoryRequests = byteCountSI(result.memoryRequests)
	result.StorageRequests = byteCountSI(result.storageRequests)

	return &result, nil
}

func byteCountSI(b float64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%.2f B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", b/float64(div), "KMGTPE"[exp])
}
