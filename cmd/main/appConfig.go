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
	"gopkg.in/alecthomas/kingpin.v2"
)

type appConfigType struct {
	Version                    string
	mode                       *string
	port                       *int
	logLevel                   *string
	kubeconfigPath             *string
	frontDist                  *string
	ingressNoFiltration        *bool
	ingressAnotationKey        *string
	ingressFilter              *string
	kubeconfigServer           *string
	gitlabURL                  *string
	gitlabToken                *string
	makeAPICallServer          *string
	registryURL                *string
	registryUser               *string
	registryPassword           *string
	registryDirectory          *string
	releasePatern              *string
	releaseNotDeleteDays       *int
	removeBranchDaysInactive   *int
	removeBranchLastScaleDate  *int
	systemGitTags              *string
	systemNamespaces           *string
	ingressHostDefaultProtocol *string
}

var appConfig = appConfigType{
	Version: gitVersion,
	mode: kingpin.Flag(
		"mode",
		"web or batch or cleanOldTags",
	).Default("web").Enum("web", "batch", "cleanOldTags"),
	port: kingpin.Flag(
		"server.port",
		"port",
	).Default("9000").Int(),
	kubeconfigPath: kingpin.Flag(
		"kubeconfig.path",
		"path to kubeconfig",
	).Default("").String(),
	logLevel: kingpin.Flag(
		"log.level",
		"logging level",
	).Default("INFO").Envar("LOG_LEVEL").String(),
	frontDist: kingpin.Flag(
		"front.dist",
		"front dist",
	).Default("front/dist").String(),
	ingressAnotationKey: kingpin.Flag(
		"ingress.anotationKey",
		"ingress anotation key",
	).Default("kubernetes-manager").String(),
	ingressNoFiltration: kingpin.Flag(
		"ingress.no-filtration",
		"ingress filter",
	).Bool(),
	ingressFilter: kingpin.Flag(
		"ingress.filter",
		"ingress filter",
	).Default("kubernetes-manager=true").String(),
	kubeconfigServer: kingpin.Flag(
		"kubeconfig.server",
		"kubeconfig server",
	).Default("https://kubernetes-api:6443").Envar("KUBERNETES_ENDPOINT").String(),
	gitlabURL: kingpin.Flag(
		"gitlab.url",
		"Gitlab api endpoint",
	).Default("https://git/api/v4").Envar("GITLAB_URL").String(),
	gitlabToken: kingpin.Flag(
		"gitlab.token",
		"Gitlab token",
	).Default("some-token").Envar("GITLAB_TOKEN").String(),
	makeAPICallServer: kingpin.Flag(
		"makeAPICall.server",
		"API server host",
	).Default("127.0.0.1").String(),
	registryURL: kingpin.Flag(
		"registry.url",
		"Docker registry url",
	).Default("http://127.0.0.1:5000").Envar("REGISTRY_URL").String(),
	registryUser: kingpin.Flag(
		"registry.user",
		"Docker registry user",
	).Default("").Envar("REGISTRY_USER").String(),
	registryPassword: kingpin.Flag(
		"registry.password",
		"Docker registry password",
	).Default("").Envar("REGISTRY_PASSWORD").String(),
	registryDirectory: kingpin.Flag(
		"registry.directory",
		"Directory with docker registry files",
	).Default("/var/lib/registry/").Envar("REGISTRY_DIRECTORY").String(),
	releasePatern: kingpin.Flag(
		"release.pattern",
		"Git tag release pattern",
	).Default(`release-(\d{4}\d{2}\d{2}).*`).String(),
	releaseNotDeleteDays: kingpin.Flag(
		"release.notDeleteDays",
		"Days to not delete git tags on release patern",
	).Default("10").Int(),
	removeBranchDaysInactive: kingpin.Flag(
		"batch.removeBranchDaysInactive",
		"Days to delete kubernetes namespace after last commit ",
	).Default("20").Int(),
	removeBranchLastScaleDate: kingpin.Flag(
		"batch.removeBranchLastScaleDate",
		"Days to delete kubernetes namespace after last scale",
	).Default("10").Int(),
	systemGitTags: kingpin.Flag(
		"system.gitTags",
		"Docker registry git tags/branch that can not delete (regexp)",
	).Default("^master$,^release-.*").Envar("SYSTEM_GIT_TAGS").String(),
	systemNamespaces: kingpin.Flag(
		"system.namespaces",
		"Kubernetes namespaces that can not delete",
	).Default("^kube-system$").Envar("SYSTEM_NAMESPACES").String(),
	ingressHostDefaultProtocol: kingpin.Flag(
		"ingressHostDefaultProtocol",
		"default host protocol",
	).Default("https").String(),
}
