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
	"bytes"
	"context"
	"io"

	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	corev1 "k8s.io/api/core/v1"
)

const defaultLogLines = 100

type GetPodContainerLogRequest struct {
	Pod        string
	Container  string
	TailLines  *int64
	Timestamps bool
}

func (l *GetPodContainerLogRequest) SetTimestamps(value string) {
	if value == "true" {
		l.Timestamps = true
	} else {
		l.Timestamps = false
	}
}

func (l *GetPodContainerLogRequest) GetTailLines() *int64 {
	if l.TailLines == nil {
		tailLines := int64(defaultLogLines)

		return &tailLines
	}

	return l.TailLines
}

func (e *Environment) GetPodContainerLog(ctx context.Context, input *GetPodContainerLogRequest) (string, error) {
	ctx, span := telemetry.Start(ctx, "api.GetPodContainerLog")
	defer span.End()

	podLogOptions := corev1.PodLogOptions{
		Container:  input.Container,
		Follow:     false,
		TailLines:  input.GetTailLines(),
		Timestamps: input.Timestamps,
	}

	podLogs, err := e.clientset.CoreV1().Pods(e.Namespace).GetLogs(input.Pod, &podLogOptions).Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
