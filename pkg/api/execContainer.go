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

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

type ExecContainerResults struct {
	Stdout   string
	Stderr   string
	ExecCode string
}

func ExecContainer(ns string, pod string, labelSelector string, container string, command string) (*ExecContainerResults, error) { //nolint:lll
	log.Debugf("container=%s,command=%s", container, command)

	clientset, err := getClientset(ns)
	if err != nil {
		return &ExecContainerResults{}, errors.Wrap(err, "can not get clientset")
	}

	namespace := getNamespace(ns)
	podName := pod

	if len(pod) == 0 {
		pods, err := clientset.CoreV1().Pods(namespace).List(Ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
			FieldSelector: runningPodSelector,
		})
		if err != nil {
			return &ExecContainerResults{}, errors.Wrap(err, "can not list pods")
		}

		podName = pods.Items[0].Name
	}

	req := clientset.CoreV1().RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   []string{"/bin/sh", "-c", command},
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	restconfig := restconfigCluster[getCluster(ns)]

	exec, err := remotecommand.NewSPDYExecutor(restconfig, "POST", req.URL())
	if err != nil {
		return &ExecContainerResults{}, errors.Wrap(err, "can not execute command")
	}

	var stdout, stderr bytes.Buffer

	err = exec.Stream(remotecommand.StreamOptions{
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
