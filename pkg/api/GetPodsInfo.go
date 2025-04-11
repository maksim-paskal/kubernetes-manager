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
	"strconv"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const podPendingDuration = 5 * time.Minute

type PodsInfo struct {
	PodsTotal       int64
	PodsReady       int64
	PodsFailed      int64
	PodsFailedName  []string
	PodsPendingName []string
	cpuRequests     float64
	memoryRequests  float64
	storageRequests float64

	CPURequests     string
	MemoryRequests  string
	StorageRequests string
}

func (e *Environment) GetPodsInfo(ctx context.Context) (*PodsInfo, error) {
	ctx, span := telemetry.Start(ctx, "api.GetPodsInfo")
	defer span.End()

	pods, err := GetCachedKubernetesPodsStatus(ctx, e.Cluster, e.Namespace, PodIsNotSucceeded)
	if err != nil {
		return nil, errors.Wrap(err, "error list pods")
	}

	telemetry.Event(span, "loaded pods", map[string]string{
		"len": strconv.Itoa(len(pods)),
	})

	result := PodsInfo{}

	result.PodsFailedName = make([]string, 0)

	for _, pod := range pods {
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
			// check if pod is still starting
			if time.Since(pod.CreationTimestamp.Time) < podPendingDuration {
				result.PodsPendingName = append(result.PodsPendingName, pod.Name)

				continue
			}

			result.PodsFailed++

			if podReason == string(corev1.PodRunning) {
				podReason = "NotReady"
			}

			podName := fmt.Sprintf("%s (%s)", pod.Name, podReason)

			result.PodsFailedName = append(result.PodsFailedName, podName)
		}
	}

	pvcs, err := GetCachedPersistentVolumeClaims(ctx, e.Cluster, e.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "error list pvc")
	}

	telemetry.Event(span, "loaded pvc", map[string]string{
		"len": strconv.Itoa(len(pvcs)),
	})

	for _, pvc := range pvcs {
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
