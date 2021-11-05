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
package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
)

var getInfoDBCommands = initPodCommands()

type getInfoDBCommandsType struct {
	param         execContainerParams
	beforeExecute func(param *execContainerParams, r *http.Request) error
	filterStdout  func(param execContainerParams, stdout string) string
}

func initPodCommands() map[string]getInfoDBCommandsType {
	m := make(map[string]getInfoDBCommandsType)

	var command strings.Builder

	command.WriteString("mongo admin -u $MONGO_INITDB_ROOT_USERNAME")
	command.WriteString(" -p $MONGO_INITDB_ROOT_PASSWORD")
	command.WriteString(" --quiet --eval  \"printjson(db.adminCommand('listDatabases'))\"")

	m["mongoInfo"] = getInfoDBCommandsType{
		param: execContainerParams{
			namespace:     "",
			labelSelector: "app=mongo",
			container:     "mongo",
			command:       command.String(),
		},
		beforeExecute: func(param *execContainerParams, r *http.Request) error {
			if err := checkParams(r, []string{"namespace"}); err != nil {
				return err
			}

			namespace := r.URL.Query()["namespace"]
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
			if err := checkParams(r, []string{"namespace", "pod"}); err != nil {
				return err
			}

			namespace := r.URL.Query()["namespace"]
			param.namespace = namespace[0]

			pod := r.URL.Query()["pod"]

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != config.KeyValueLength {
				return errNoPodSelected
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
			if err := checkParams(r, []string{"namespace", "pod"}); err != nil {
				return err
			}

			namespace := r.URL.Query()["namespace"]
			param.namespace = namespace[0]

			pod := r.URL.Query()["pod"]

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != config.KeyValueLength {
				return errNoPodSelected
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
			if err := checkParams(r, []string{"namespace", "pod"}); err != nil {
				return err
			}

			namespace := r.URL.Query()["namespace"]
			param.namespace = namespace[0]

			pod := r.URL.Query()["pod"]

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != config.KeyValueLength {
				return errNoPodSelected
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
			if err := checkParams(r, []string{"namespace", "text", "pod"}); err != nil {
				return err
			}

			namespace := r.URL.Query()["namespace"]
			text := r.URL.Query()["text"]

			param.namespace = namespace[0]
			param.command = fmt.Sprintf("%s %s", param.command, text)

			pod := r.URL.Query()["pod"]

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != config.KeyValueLength {
				return errNoPodSelected
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
			if err := checkParams(r, []string{"namespace", "pod"}); err != nil {
				return err
			}

			namespace := r.URL.Query()["namespace"]
			param.namespace = namespace[0]

			pod := r.URL.Query()["pod"]

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != config.KeyValueLength {
				return errNoPodSelected
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
			if err := checkParams(r, []string{"namespace", "origin", "branch", "pod"}); err != nil {
				return err
			}

			namespace := r.URL.Query()["namespace"]
			origin := r.URL.Query()["origin"]
			branch := r.URL.Query()["branch"]

			param.namespace = namespace[0]
			param.command = fmt.Sprintf("%s %s %s", param.command, origin[0], branch[0])

			pod := r.URL.Query()["pod"]
			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != config.KeyValueLength {
				return errNoPodSelected
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
			if err := checkParams(r, []string{"namespace", "pod"}); err != nil {
				return err
			}

			namespace := r.URL.Query()["namespace"]
			param.namespace = namespace[0]

			pod := r.URL.Query()["pod"]

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != config.KeyValueLength {
				return errNoPodSelected
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
			if err := checkParams(r, []string{"namespace", "pod"}); err != nil {
				return err
			}

			namespace := r.URL.Query()["namespace"]

			param.namespace = namespace[0]
			pod := r.URL.Query()["pod"]

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != config.KeyValueLength {
				return errNoPodSelected
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
			if err := checkParams(r, []string{"namespace", "pod"}); err != nil {
				return err
			}

			namespace := r.URL.Query()["namespace"]
			param.namespace = namespace[0]
			pod := r.URL.Query()["pod"]

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != config.KeyValueLength {
				return errNoPodSelected
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
			if err := checkParams(r, []string{"namespace", "pod"}); err != nil {
				return err
			}
			namespace := r.URL.Query()["namespace"]

			param.namespace = namespace[0]
			pod := r.URL.Query()["pod"]

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != config.KeyValueLength {
				return errNoPodSelected
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
			if err := checkParams(r, []string{"namespace", "pod"}); err != nil {
				return err
			}

			namespace := r.URL.Query()["namespace"]
			param.namespace = namespace[0]
			pod := r.URL.Query()["pod"]

			podinfo := strings.Split(pod[0], ":")

			if len(podinfo) != config.KeyValueLength {
				return errNoPodSelected
			}

			param.podname = podinfo[0]
			param.container = podinfo[1]

			return nil
		},
	}

	return m
}
