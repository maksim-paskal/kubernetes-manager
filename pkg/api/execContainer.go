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

	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

type ExecContainerResults struct {
	Stdout   string
	Stderr   string
	ExecCode string
}

// exec command in container
// @container must contains <pod>:<container>.
func (e *Environment) ExecContainer(ctx context.Context, container string, command string) (*ExecContainerResults, error) {
	ctx, span := telemetry.Start(ctx, "api.ExecContainer")
	defer span.End()

	containerInfo, err := types.NewContainerInfo(container)
	if err != nil {
		return nil, err
	}

	req := e.clientset.CoreV1().RESTClient().
		Post().
		Namespace(e.Namespace).
		Resource("pods").
		Name(containerInfo.PodName).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerInfo.ContainerName,
			Command:   []string{"/bin/sh", "-c", command},
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	restconfig, err := client.GetRestConfig(e.Cluster)
	if err != nil {
		return nil, errors.New("can not get client config for cluster")
	}

	exec, err := remotecommand.NewSPDYExecutor(restconfig, "POST", req.URL())
	if err != nil {
		return &ExecContainerResults{}, errors.Wrap(err, "can not execute command")
	}

	var stdout, stderr bytes.Buffer

	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	results := ExecContainerResults{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err != nil {
		results.ExecCode = err.Error()
	}

	return &results, nil
}
