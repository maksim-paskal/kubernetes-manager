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
	"gopkg.in/alecthomas/kingpin.v2"
)

type appConfigType struct {
	Version             string
	mode                *string
	port                *int
	kubeconfigPath      *string
	frontDist           *string
	ingressNoFiltration *bool
	ingressAnotationKey *string
	ingressFilter       *string
	kubeconfigServer    *string
	gitlabURL           *string
	gitlabToken         *string
	makeAPICallServer   *string
	namespacesNotDelete *string
	logLevel            *string
	logPrety            *bool
}

var appConfig = appConfigType{
	Version: "1.0.1",
	mode: kingpin.Flag(
		"mode",
		"execute mode",
	).Default("web").Envar("KM_MODE").Enum("web", "batch", "cleanOldTags", "pauseALL"),
	port: kingpin.Flag(
		"server.port",
		"port",
	).Default("9000").Envar("KM_SERVER_PORT").Int(),
	kubeconfigPath: kingpin.Flag(
		"kubeconfig.path",
		"path to kubeconfig",
	).Default("").Envar("KM_KUBECONFIG_PATH").String(),
	frontDist: kingpin.Flag(
		"front.dist",
		"front dist",
	).Default("front/dist").Envar("KM_FRONT_DIST").String(),
	ingressAnotationKey: kingpin.Flag(
		"ingress.anotationKey",
		"ingress anotation key",
	).Default("kubernetes-manager").Envar("KM_INGRESS_ANOTATION_KEY").String(),
	ingressNoFiltration: kingpin.Flag(
		"ingress.no-filtration",
		"ingress filter",
	).Bool(),
	ingressFilter: kingpin.Flag(
		"ingress.filter",
		"ingress filter",
	).Default("kubernetes-manager=true").Envar("KM_INGRESS_FILTER").String(),
	kubeconfigServer: kingpin.Flag(
		"kubeconfig.server",
		"kubeconfig server",
	).Default("https://kubernetes-api-endpoint:6443").Envar("KM_KUBECONFIG_SERVER").String(),
	gitlabURL: kingpin.Flag(
		"gitlab.url",
		"url to api",
	).Default("https://gitlab-server/api/v4").Envar("KM_GITLAB_URL").String(),
	gitlabToken: kingpin.Flag(
		"gitlab.token",
		"token to api",
	).Default("gitlab-token").Envar("KM_GITLAB_TOKEN").String(),
	makeAPICallServer: kingpin.Flag(
		"makeAPICall.server",
		"server for api call",
	).Default("127.0.0.1").Envar("KM_MAKEAPICALL_SERVER").String(),
	namespacesNotDelete: kingpin.Flag(
		"namespaces.not-deleted",
		"namespaces that not be deleted, comma separeted",
	).Default("master").Envar("KM_NAMESPACES_NOT_DELETED").String(),
	logLevel: kingpin.Flag(
		"log.level",
		"logger level",
	).Default("INFO").Envar("LOG_LEVEL").String(),
	logPrety: kingpin.Flag(
		"prety",
		"logs prety",
	).Bool(),
}
