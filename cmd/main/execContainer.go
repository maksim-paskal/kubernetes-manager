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
package main

import (
	"bytes"
	"context"

	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/deprecated/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

type execContainerParams struct {
	namespace     string
	labelSelector string
	podname       string
	container     string
	command       string
}

type execContainerResults struct {
	Stdout   string
	Stderr   string
	ExecCode string
}

func execContainer(rootSpan opentracing.Span, params execContainerParams) (execContainerResults, error) {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("execContainer", opentracing.ChildOf(rootSpan.Context()))

	defer span.Finish()

	span.SetTag("params", params)

	if len(params.podname) == 0 {
		span.LogKV("event", "pod list start")

		pods, err := clientset.CoreV1().Pods(params.namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: params.labelSelector,
			FieldSelector: "status.phase=Running",
		})

		span.LogKV("event", "pod list end")

		if err != nil {
			log.
				WithError(err).
				WithField(logrushookopentracing.SpanKey, span).
				Error()

			return execContainerResults{}, errors.Wrap(err, "error list pods")
		}

		if len(pods.Items) == 0 {
			log.
				WithError(ErrNoPodInStatusRunning).
				WithField(logrushookopentracing.SpanKey, span).
				Error()

			return execContainerResults{}, ErrNoPodInStatusRunning
		}

		params.podname = pods.Items[0].Name
	}

	req := clientset.CoreV1().RESTClient().
		Post().
		Namespace(params.namespace).
		Resource("pods").
		Name(params.podname).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: params.container,
			Command:   []string{"/bin/sh", "-c", params.command},
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	span.LogKV("event", "remotecommand start")

	exec, err := remotecommand.NewSPDYExecutor(restconfig, "POST", req.URL())
	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			Error()

		return execContainerResults{}, errors.Wrap(err, "error in NewSPDYExecutor")
	}

	span.LogKV("event", "remotecommand end")

	var stdout, stderr bytes.Buffer

	span.LogKV("event", "stream start")

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	span.LogKV("event", "stream end")

	results := execContainerResults{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err != nil {
		results.ExecCode = err.Error()
	}

	return results, nil
}
