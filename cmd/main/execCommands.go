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
	"encoding/json"
	"net/http"

	sentry "github.com/getsentry/sentry-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func execCommands(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
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
	if err != nil { //nolint:wsl
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")

		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)

	if err != nil {
		logError(span, sentry.LevelError, r, err, "")
	}
}
