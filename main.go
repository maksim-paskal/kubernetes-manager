/*
Copyright paskal.maksim@gmail.com
Licensed under the Apache License, Version 2.0 (the "License");
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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	sentry "github.com/getsentry/sentry-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log "github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/alecthomas/kingpin.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

var (
	buildTime string
)

var getInfoDBCommands = initPodCommands()

type getInfoDBCommandsType struct {
	param         execContainerParams
	beforeExecute func(param *execContainerParams, r *http.Request) error
	filterStdout  func(param execContainerParams, stdout string) string
}

func initPodCommands() map[string]getInfoDBCommandsType {
	m := make(map[string]getInfoDBCommandsType)

	m["mongoInfo"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace:     "",
			labelSelector: "app=mongo",
			container:     "mongo",
			command:       "mongo admin -u $MONGO_INITDB_ROOT_USERNAME -p $MONGO_INITDB_ROOT_PASSWORD --quiet --eval  \"printjson(db.adminCommand('listDatabases'))\"",
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			namespace := r.URL.Query()["namespace"]

			if len(namespace) != 1 {
				return errors.New("no namespace")
			}
			param.namespace = namespace[0]
			return nil
		},
	}
	m["mongoMigrations"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace: "",
			podname:   "",
			container: "",
			command:   "/kubernetes-manager/mongoMigrations",
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			namespace := r.URL.Query()["namespace"]

			if len(namespace) != 1 {
				return errors.New("no namespace")
			}
			param.namespace = namespace[0]

			pod := r.URL.Query()["pod"]

			if len(pod) != 1 {
				return errors.New("no pod")
			}

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != 2 {
				return errors.New("no pod selected")
			}

			param.podname = podinfo[0]
			param.container = podinfo[1]
			return nil
		},
	}
	m["xdebugInfo"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace: "",
			podname:   "",
			container: "",
			command:   "/kubernetes-manager/xdebugInfo",
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			namespace := r.URL.Query()["namespace"]

			if len(namespace) != 1 {
				return errors.New("no namespace")
			}
			param.namespace = namespace[0]

			pod := r.URL.Query()["pod"]

			if len(pod) != 1 {
				return errors.New("no pod")
			}

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != 2 {
				return errors.New("no pod selected")
			}

			param.podname = podinfo[0]
			param.container = podinfo[1]

			return nil
		},
	}
	m["xdebugEnable"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace: "",
			podname:   "",
			container: "",
			command:   "/kubernetes-manager/enableXdebug",
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			namespace := r.URL.Query()["namespace"]

			if len(namespace) != 1 {
				return errors.New("no namespace")
			}
			param.namespace = namespace[0]

			pod := r.URL.Query()["pod"]

			if len(pod) != 1 {
				return errors.New("no pod")
			}

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != 2 {
				return errors.New("no pod selected")
			}

			param.podname = podinfo[0]
			param.container = podinfo[1]

			return nil
		},
	}
	m["setPhpSettings"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace: "",
			podname:   "",
			container: "",
			command:   "/kubernetes-manager/setPhpSettings",
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			namespace := r.URL.Query()["namespace"]

			if len(namespace) != 1 {
				return errors.New("no namespace")
			}

			text := r.URL.Query()["text"]

			if len(text) != 1 {
				return errors.New("no text")
			}
			param.namespace = namespace[0]
			param.command = fmt.Sprintf("%s %s", param.command, text)

			pod := r.URL.Query()["pod"]

			if len(pod) != 1 {
				return errors.New("no pod")
			}

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != 2 {
				return errors.New("no pod selected")
			}

			param.podname = podinfo[0]
			param.container = podinfo[1]

			return nil
		},
	}
	m["getPhpSettings"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace: "",
			podname:   "",
			container: "",
			command:   "/kubernetes-manager/getPhpSettings",
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			namespace := r.URL.Query()["namespace"]

			if len(namespace) != 1 {
				return errors.New("no namespace")
			}

			param.namespace = namespace[0]

			pod := r.URL.Query()["pod"]

			if len(pod) != 1 {
				return errors.New("no pod")
			}

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != 2 {
				return errors.New("no pod selected")
			}

			param.podname = podinfo[0]
			param.container = podinfo[1]

			return nil
		},
	}
	m["enableGit"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace: "",
			podname:   "",
			container: "",
			command:   "/kubernetes-manager/enableGit",
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			namespace := r.URL.Query()["namespace"]

			if len(namespace) != 1 {
				return errors.New("no namespace")
			}

			origin := r.URL.Query()["origin"]

			if len(origin) != 1 {
				return errors.New("no origin")
			}

			branch := r.URL.Query()["branch"]

			if len(origin) != 1 {
				return errors.New("no branch")
			}

			param.namespace = namespace[0]
			param.command = fmt.Sprintf("%s %s %s", param.command, origin[0], branch[0])

			pod := r.URL.Query()["pod"]

			if len(pod) != 1 {
				return errors.New("no pod")
			}

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != 2 {
				return errors.New("no pod selected")
			}

			param.podname = podinfo[0]
			param.container = podinfo[1]

			return nil
		},
	}
	m["getGitPubKey"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace: "",
			podname:   "",
			container: "",
			command:   "/kubernetes-manager/getGitPubKey",
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			namespace := r.URL.Query()["namespace"]

			if len(namespace) != 1 {
				return errors.New("no namespace")
			}

			param.namespace = namespace[0]

			pod := r.URL.Query()["pod"]

			if len(pod) != 1 {
				return errors.New("no pod")
			}

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != 2 {
				return errors.New("no pod selected")
			}

			param.podname = podinfo[0]
			param.container = podinfo[1]
			return nil
		},
	}
	m["gitFetch"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace: "",
			podname:   "",
			container: "",
			command:   "/kubernetes-manager/gitFetch",
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			namespace := r.URL.Query()["namespace"]

			if len(namespace) != 1 {
				return errors.New("no namespace")
			}

			param.namespace = namespace[0]
			pod := r.URL.Query()["pod"]

			if len(pod) != 1 {
				return errors.New("no pod")
			}

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != 2 {
				return errors.New("no pod selected")
			}

			param.podname = podinfo[0]
			param.container = podinfo[1]
			return nil
		},
	}
	m["clearCache"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace: "",
			podname:   "",
			container: "",
			command:   "/kubernetes-manager/clearCache",
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			namespace := r.URL.Query()["namespace"]

			if len(namespace) != 1 {
				return errors.New("no namespace")
			}
			param.namespace = namespace[0]
			pod := r.URL.Query()["pod"]

			if len(pod) != 1 {
				return errors.New("no pod")
			}

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != 2 {
				return errors.New("no pod selected")
			}

			param.podname = podinfo[0]
			param.container = podinfo[1]
			return nil
		},
	}
	m["getGitBranch"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace: "",
			podname:   "",
			container: "",
			command:   "/kubernetes-manager/getGitBranch",
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			namespace := r.URL.Query()["namespace"]

			if len(namespace) != 1 {
				return errors.New("no namespace")
			}

			param.namespace = namespace[0]
			pod := r.URL.Query()["pod"]

			if len(pod) != 1 {
				return errors.New("no pod")
			}

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != 2 {
				return errors.New("no pod selected")
			}

			param.podname = podinfo[0]
			param.container = podinfo[1]
			return nil
		},
	}
	m["mysqlMigrations"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace: "",
			podname:   "",
			container: "",
			command:   "/kubernetes-manager/mysqlMigrations",
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			namespace := r.URL.Query()["namespace"]

			if len(namespace) != 1 {
				return errors.New("no namespace")
			}
			param.namespace = namespace[0]
			pod := r.URL.Query()["pod"]

			if len(pod) != 1 {
				return errors.New("no pod")
			}

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != 2 {
				return errors.New("no pod selected")
			}

			param.podname = podinfo[0]
			param.container = podinfo[1]
			return nil
		},
	}
	return m
}

type httpResponse struct {
	Status string
	Body   string
}

func deleteALL(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("deleteALL", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	namespace := r.URL.Query()["namespace"]

	if len(namespace) != 1 {
		http.Error(w, "namespace not set", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "namespace not set")
		return
	}

	if stringInSlice(namespace[0], strings.Split(*appConfig.namespacesNotDelete, ",")) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{status:'ok',warning:'namespace can not be deleted'}"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelInfo, r, err, "")
		}
		return
	}

	logError(span, sentry.LevelDebug, r, nil, "user requested deleteALL")

	type ResultData struct {
		DeleteNamespaceResultBody   httpResponse
		DeleteRegistryTagResultBody httpResponse
	}
	type ResultType struct {
		Result ResultData `json:"result"`
	}
	result := ResultType{
		Result: ResultData{},
	}

	var q url.Values

	ch3 := make(chan httpResponse)
	//ch3 = make(chan httpResponse)

	q = make(url.Values)
	q.Add("namespace", namespace[0])
	go makeAPICall(span, "/api/deleteNamespace", q, ch3)

	result.Result.DeleteNamespaceResultBody = (<-ch3)

	projectID := r.URL.Query()["git-project-id"]
	tag := r.URL.Query()["registry-tag"]

	if len(projectID) == 1 && len(tag) == 1 {
		ch4 := make(chan httpResponse)
		q = make(url.Values)
		q.Add("projectID", r.URL.Query()["git-project-id"][0])
		q.Add("tag", r.URL.Query()["registry-tag"][0])
		go makeAPICall(span, "/api/deleteRegistryTag", q, ch4)

		result.Result.DeleteRegistryTagResultBody = (<-ch4)
	} else {
		result.Result.DeleteRegistryTagResultBody = httpResponse{
			Status: "not executed",
			Body:   "projectID or tag not set",
		}
	}

	span.LogKV("result", result)
	js, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, err, "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, err, "")
	}
}

func execCommands(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("execCommands", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	cmd := r.URL.Query()["cmd"]

	if len(cmd) != 1 {
		http.Error(w, "no command", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "no command")
		return
	}

	_, ok := getInfoDBCommands[cmd[0]]
	if !ok {
		http.Error(w, "no command found", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "no command found")
		return
	}

	podExecute := getInfoDBCommands[cmd[0]]

	if podExecute.beforeExecute != nil {
		err := podExecute.beforeExecute(&podExecute.param, r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelError, r, err, "")
			return
		}
	}

	execResults, err := execContainer(span, podExecute.param)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}

	type ResultType struct {
		Result execContainerResults `json:"result"`
	}

	if podExecute.filterStdout != nil {
		execResults.Stdout = podExecute.filterStdout(podExecute.param, execResults.Stdout)
	}

	result := ResultType{
		Result: execResults,
	}

	span.LogKV("result", result)
	js, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, err, "")
	}
}

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
	var tracer = opentracing.GlobalTracer()
	span := tracer.StartSpan("execContainer", opentracing.ChildOf(rootSpan.Context()))
	defer span.Finish()

	span.SetTag("params", params)

	if len(params.podname) == 0 {
		span.LogKV("event", "pod list start")
		pods, err := clientset.CoreV1().Pods(params.namespace).List(metav1.ListOptions{
			LabelSelector: params.labelSelector,
			FieldSelector: "status.phase=Running",
		})
		span.LogKV("event", "pod list end")

		if err != nil {
			logError(span, sentry.LevelError, nil, err, "")

			return execContainerResults{}, err
		}

		if len(pods.Items) == 0 {
			logError(span, sentry.LevelInfo, nil, nil, "pod in status Running not found, retry")
			return execContainerResults{}, errors.New("pod in status Running not found, retry")
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
		logError(span, sentry.LevelError, nil, err, "")

		return execContainerResults{}, err
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
func getNamespace(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getNamespace", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	namespace := r.URL.Query()["namespace"]

	if len(namespace) < 1 {
		http.Error(w, "namespace not set", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "namespace not set")
		return
	}

	_, err := clientset.CoreV1().Namespaces().Get(namespace[0], metav1.GetOptions{})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte("{status:'ok'}"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, err, "")
	}
}
func deleteRegistryTag(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("deleteRegistryTag", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	tag := r.URL.Query()["tag"]

	if len(tag) < 1 {
		http.Error(w, "tag not set", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "tag not set")
		return
	}

	if strings.EqualFold(tag[0], "master") || strings.EqualFold(tag[0], "pp") || strings.HasPrefix(tag[0], "release-") {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{status:'ok',warning:'registry tag can not be deleted'}"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelInfo, r, err, "")
		}
		return
	}

	projectID := r.URL.Query()["projectID"]

	if len(projectID) < 1 {
		http.Error(w, "projectID not set", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "projectID not set")
		return
	}

	span.LogKV("params", fmt.Sprintf("projectID=%s,tag=%s", projectID[0], tag[0]))

	git := gitlab.NewClient(nil, *appConfig.gitlabToken)
	err := git.SetBaseURL(*appConfig.gitlabURL)
	if err != nil {
		log.Panic(err)
	}
	span.LogKV("event", "ListRegistryRepositories")
	gitRepos, _, err := git.ContainerRegistry.ListRegistryRepositories(projectID[0], nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}

	for _, gitRepo := range gitRepos {
		span.LogKV("DeleteRegistryRepositoryTag", fmt.Sprintf("gitRepo.ID=%d", gitRepo.ID))

		_, err := git.ContainerRegistry.DeleteRegistryRepositoryTag(projectID[0], gitRepo.ID, tag[0])

		if err != nil {
			span.LogKV("warning", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte("{status:'ok'}"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, err, "")
	}
}
func executeBatch(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("executeBatch", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	batch(span)

	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte("{status:'ok'}"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, err, "")
	}
}
func deleteNamespace(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("deleteNamespace", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	namespace := r.URL.Query()["namespace"]

	if len(namespace) < 1 {
		http.Error(w, "namespace not set", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "namespace not set")
		return
	}

	if stringInSlice(namespace[0], strings.Split(*appConfig.namespacesNotDelete, ",")) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{status:'ok',warning:'namespace can not be deleted'}"))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelError, r, err, "")
		}
		return
	}

	err := clientset.CoreV1().Namespaces().Delete(namespace[0], nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte("{status:'ok'}"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
	}

}
func deletePod(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("deletePod", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	namespace := r.URL.Query()["namespace"]

	if len(namespace) < 1 {
		http.Error(w, "namespace not set", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "namespace not set")
		return
	}

	GracePeriodSeconds := int64(0)

	opt := &metav1.DeleteOptions{
		GracePeriodSeconds: &GracePeriodSeconds,
	}

	podName := ""
	LabelSelector := r.URL.Query()["LabelSelector"]
	pod := r.URL.Query()["pod"]

	if len(pod) > 0 {
		podinfo := strings.Split(pod[0], ":")

		if len(podinfo) != 2 {
			http.Error(w, "no pod selected", http.StatusInternalServerError)
			logError(span, sentry.LevelInfo, r, nil, "no pod selected")
			return
		}

		podName = podinfo[0]
	} else {
		if len(LabelSelector) < 1 {
			http.Error(w, "LabelSelector not set", http.StatusInternalServerError)
			logError(span, sentry.LevelInfo, r, nil, "LabelSelector not set")
			return
		}

		pods, err1 := clientset.CoreV1().Pods(namespace[0]).List(metav1.ListOptions{
			LabelSelector: LabelSelector[0],
			FieldSelector: "status.phase=Running",
		})

		if err1 != nil {
			http.Error(w, err1.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelError, r, err1, "")
			return
		}

		if len(pods.Items) == 0 {
			http.Error(w, "pod in status Running not found, retry", http.StatusInternalServerError)
			logError(span, sentry.LevelInfo, r, nil, "pod in status Running not found, retry")
			return
		}

		podName = pods.Items[0].Name
	}
	err2 := clientset.CoreV1().Pods(namespace[0]).Delete(podName, opt)

	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err2, "")
		return
	}

	type ResultData struct {
		Stdout string
	}

	type ResultType struct {
		Result ResultData `json:"result"`
	}
	result := ResultType{
		Result: ResultData{
			Stdout: fmt.Sprintf("deleted %s pod", podName),
		},
	}

	js, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
	}
}

func getRunningPodsCount(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getRunningPodsCount", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	namespace := r.URL.Query()["namespace"]

	if len(namespace) < 1 {
		http.Error(w, "namespace not set", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "namespace not set")
		return
	}

	pods, err := clientset.CoreV1().Pods(namespace[0]).List(metav1.ListOptions{
		FieldSelector: "status.phase=Running",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "max-age=10")

	_, err = w.Write([]byte(fmt.Sprintf("{\"count\":%d}", len(pods.Items))))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
	}
}

func getAPIversion(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getAPIversion", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(fmt.Sprintf("{\"version\":\"%s-%s\"}", appConfig.Version, buildTime)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
	}
}

func getPods(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getPods", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	namespace := r.URL.Query()["namespace"]

	if len(namespace) < 1 {
		http.Error(w, "namespace not set", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "namespace not set")
		return
	}

	pods, err := clientset.CoreV1().Pods(namespace[0]).List(metav1.ListOptions{
		FieldSelector: "status.phase=Running",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}

	if len(pods.Items) == 0 {
		http.Error(w, "pod in status Running not found, retry", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "pod in status Running not found, retry")
		return
	}

	type PodContainerData struct {
		ContainerName string
	}

	type PodData struct {
		PodName       string
		PodLabels     map[string]string
		PodContainers []PodContainerData
	}

	type ResultType struct {
		Result []PodData `json:"result"`
	}

	var podsData []PodData

	for _, pod := range pods.Items {
		var podContainersData []PodContainerData

		for _, podContainer := range pod.Spec.Containers {
			podContainerData := PodContainerData{
				ContainerName: podContainer.Name,
			}

			podContainersData = append(podContainersData, podContainerData)
		}
		podData := PodData{
			PodName:       pod.Name,
			PodLabels:     pod.Labels,
			PodContainers: podContainersData,
		}
		podsData = append(podsData, podData)
	}

	result := ResultType{
		Result: podsData,
	}

	js, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
	}
}

func getIngress(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getIngress", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	type IngressList struct {
		Namespace            string
		NamespaceStatus      string
		NamespaceCreated     string
		NamespaceCreatedDays int
		IngressName          string
		IngressAnotations    map[string]string
		IngressLabels        map[string]string
		Hosts                []string
		RunningPodsCount     int
	}

	type IngressListResult struct {
		Result []IngressList `json:"result"`
	}

	var result IngressListResult

	opt := metav1.ListOptions{
		LabelSelector: *appConfig.ingressFilter,
	}
	if *appConfig.ingressNoFiltration {
		opt = metav1.ListOptions{}
	}

	span.LogKV("event", "start ingress list")
	ingresss, err := clientset.ExtensionsV1beta1().Ingresses("").List(opt)
	span.LogKV("event", "end ingress list")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}

	span.LogKV("event", "start range")
	for _, ingress := range ingresss.Items {
		var item IngressList

		span.LogKV("event", "search namespace="+ingress.Namespace)
		namespace, err := clientset.CoreV1().Namespaces().Get(ingress.Namespace, metav1.GetOptions{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelError, r, err, "")
			return
		}

		item.Namespace = namespace.Name
		item.NamespaceStatus = string(namespace.Status.Phase)
		item.NamespaceCreated = namespace.CreationTimestamp.String()
		item.RunningPodsCount = -1
		t1 := time.Now()
		t2 := namespace.CreationTimestamp.Time
		diff := t1.Sub(t2).Hours() / 24
		item.NamespaceCreatedDays = int(diff)

		item.IngressName = ingress.Name
		item.IngressAnotations = ingress.Annotations
		item.IngressLabels = ingress.Labels

		for _, rule := range ingress.Spec.Rules {
			host := "http://" + rule.Host
			if !stringInSlice(host, item.Hosts) {
				item.Hosts = append(item.Hosts, host)
			}
		}

		if len(item.Hosts) > 0 {
			result.Result = append(result.Result, item)
		}
	}
	span.LogKV("event", "end range")
	span.LogKV("result", result)
	js, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "max-age=10")
	_, err = w.Write(js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
	}
}

var clientset *kubernetes.Clientset
var restconfig *rest.Config

func scaleNamespace(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("scaleNamespace", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	namespace := r.URL.Query()["namespace"]
	version := 0

	if len(namespace) < 1 {
		http.Error(w, "namespace not set", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "namespace not set")
		return
	}

	if len(r.URL.Query()["version"]) == 1 {
		var err error

		version, err = strconv.Atoi(r.URL.Query()["version"][0])
		if err != nil {
			logError(span, sentry.LevelWarning, r, nil, fmt.Sprintf("can not parse version %s", r.URL.Query()["version"][0]))
		}
	}

	replicas := r.URL.Query()["replicas"]

	if len(replicas) < 1 {
		http.Error(w, "replicas not set", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "replicas not set")
		return
	}

	ds, err := clientset.AppsV1().Deployments(namespace[0]).List(metav1.ListOptions{})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}
	for _, d := range ds.Items {
		dps, err := clientset.AppsV1().Deployments(namespace[0]).Get(d.Name, metav1.GetOptions{})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelError, r, err, "")
			return
		}

		i, err := strconv.ParseInt(replicas[0], 10, 32)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelError, r, err, "")
			return
		}

		i32 := int32(i)
		dps.Spec.Replicas = &i32
		_, errUpdate := clientset.AppsV1().Deployments(namespace[0]).Update(dps)

		if errUpdate != nil {
			http.Error(w, errUpdate.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelError, r, errUpdate, "")
			return
		}
	}

	// scale statefullsets
	if version > 0 {
		sf, err := clientset.AppsV1().StatefulSets(namespace[0]).List(metav1.ListOptions{})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelError, r, err, "")
			return
		}

		for _, s := range sf.Items {
			ss, err := clientset.AppsV1().StatefulSets(namespace[0]).Get(s.Name, metav1.GetOptions{})

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logError(span, sentry.LevelError, r, err, "")
				return
			}

			i, err := strconv.ParseInt(replicas[0], 10, 32)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logError(span, sentry.LevelError, r, err, "")
				return
			}

			i32 := int32(i)
			ss.Spec.Replicas = &i32
			_, errUpdate := clientset.AppsV1().StatefulSets(namespace[0]).Update(ss)

			if errUpdate != nil {
				http.Error(w, errUpdate.Error(), http.StatusInternalServerError)
				logError(span, sentry.LevelError, r, errUpdate, "")
				return
			}
		}
	}
	type ResultData struct {
		Stdout string
	}

	type ResultType struct {
		Result ResultData `json:"result"`
	}
	result := ResultType{
		Result: ResultData{
			Stdout: fmt.Sprintf("all Deployments/StatefulSets in namespace scaled to %s", replicas[0]),
		},
	}
	js, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
	}
}

func addCacheControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "max-age=31557600")
		h.ServeHTTP(w, r)
	})
}

func main() {
	log.Infof("Starting kubernetes-manager %s-%s", appConfig.Version, buildTime)
	kingpin.Version(appConfig.Version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	var err error

	if !*appConfig.logPrety {
		log.SetFormatter(&log.JSONFormatter{})
	}
	logLevel, err := log.ParseLevel(*appConfig.logLevel)
	if err != nil {
		log.Panic(err)
	}
	log.SetLevel(logLevel)
	if logLevel == log.DebugLevel {
		log.SetReportCaller(true)
	}

	if len(os.Getenv("SENTRY_DSN")) > 0 {
		log.Debug("Use Sentry logging...")
		err = sentry.Init(sentry.ClientOptions{
			Release: fmt.Sprintf("%s-%s", appConfig.Version, buildTime),
		})

		if err != nil {
			fmt.Printf("Sentry initialization failed: %v\n", err)
		}
	}

	if len(*appConfig.kubeconfigPath) > 0 {
		restconfig, err = clientcmd.BuildConfigFromFlags("", *appConfig.kubeconfigPath)
		if err != nil {
			sentry.CaptureException(err)
			sentry.Flush(time.Second)

			log.Panic(err.Error())
		}
	} else {
		log.Info("No kubeconfig file use incluster")
		restconfig, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	clientset, err = kubernetes.NewForConfig(restconfig)
	if err != nil {
		sentry.CaptureException(err)
		sentry.Flush(time.Second)

		log.Panic(err.Error())
	}

	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		sentry.CaptureException(err)
		sentry.Flush(time.Second)

		log.Panicf("Could not parse Jaeger env vars: %s", err.Error())
	}
	cfg.ServiceName = "kubernetes-manager"
	cfg.Sampler.Type = jaeger.SamplerTypeConst
	cfg.Sampler.Param = 1

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		sentry.CaptureException(err)
		sentry.Flush(time.Second)

		log.Panicf("Could not initialize jaeger tracer: %s", err.Error())
	}
	defer closer.Close()

	if *appConfig.mode == "batch" {
		span := tracer.StartSpan("main")
		defer span.Finish()

		batch(span)
		return
	}

	if *appConfig.mode == "cleanOldTags" {
		span := tracer.StartSpan("main")
		defer span.Finish()

		cleanOldTags(span)
		return
	}

	if *appConfig.mode == "pauseALL" {
		span := tracer.StartSpan("main")
		defer span.Finish()

		pauseALL(span)
		return
	}

	log.Info(fmt.Sprintf("Starting on port %d...", *appConfig.port))
	fs := http.FileServer(http.Dir(*appConfig.frontDist))

	http.Handle("/", fs)
	http.Handle("/_nuxt/", addCacheControl(fs))
	http.HandleFunc("/api/getIngress", getIngress)
	http.HandleFunc("/api/getNamespace", getNamespace)
	http.HandleFunc("/api/deleteNamespace", deleteNamespace)
	http.HandleFunc("/api/deleteRegistryTag", deleteRegistryTag)
	http.HandleFunc("/api/deletePod", deletePod)
	http.HandleFunc("/api/exec", execCommands)
	http.HandleFunc("/api/deleteALL", deleteALL)
	http.HandleFunc("/api/executeBatch", executeBatch)
	http.HandleFunc("/getKubeConfig", getKubeConfig)
	http.HandleFunc("/api/scaleNamespace", scaleNamespace)
	http.HandleFunc("/api/getRunningPodsCount", getRunningPodsCount)
	http.HandleFunc("/api/version", getAPIversion)
	http.HandleFunc("/api/getPods", getPods)

	err = http.ListenAndServe(fmt.Sprintf(":%d", *appConfig.port), nil)
	if err != nil {

		sentry.CaptureException(err)
		sentry.Flush(time.Second)

		log.Fatal("ListenAndServe: ", err)
	}
}

func getKubeConfig(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getKubeConfig", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	caCRT, err := ioutil.ReadFile("/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}

	token, err := ioutil.ReadFile("/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}

	var kubeConfig = `apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: {{ .ClusterServer }}
  name: kubernetes-manager
contexts:
- context:
    cluster: kubernetes-manager
    user: kubernetes-manager
  name: kubernetes-manager
current-context: kubernetes-manager
kind: Config
preferences: {}
users:
- name: kubernetes-manager
  user:
    token: {{ .UserToken }}`

	type Inventory struct {
		ClusterCAD    string
		ClusterServer string
		UserToken     string
	}
	params := Inventory{
		ClusterCAD:    base64.StdEncoding.EncodeToString(caCRT),
		ClusterServer: *appConfig.kubeconfigServer,
		UserToken:     string(token),
	}
	var out bytes.Buffer
	tmpl, err := template.New("kubeconfig").Parse(kubeConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}
	err = tmpl.Execute(&out, params)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, err, "")
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=\"kubeconfig\"")
	_, err = w.Write(out.Bytes())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
	}
}
