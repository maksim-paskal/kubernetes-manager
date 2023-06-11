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
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/metrics"
	"github.com/maksim-paskal/kubernetes-manager/pkg/modules/autotests"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	clickRefreshButton   = "Pipeline(s) successfully created. Click Refresh button to see status."
	scaleMaxTime         = 5 * time.Minute
	noContainerSpecified = "no container specified"
)

func handlerEnvironment(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("environmentHandler", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	vars := mux.Vars(r)

	result, err := environmentOperation(r.Context(), r, vars["environmentID"], vars["operation"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.WithError(err).Error()
		}

		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	if result.cached {
		w.Header().Set("Cache-Control", "max-age=10")
	}

	for key, value := range result.headers {
		w.Header().Set(key, value)
	}

	if result.output == HandlerResultOutputJSON {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.WithError(err).Error()
		}
	} else {
		w.Header().Set("Content-Type", "text/plain")

		resultRAW, ok := result.Result.([]byte)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("can not convert result to []byte"))
			if err != nil {
				log.WithError(err).Error()
			}

			return
		}

		if _, err := w.Write(resultRAW); err != nil {
			log.WithError(err).Error()
		}
	}
}

func environmentOperation(ctx context.Context, r *http.Request, environmentID string, operation string) (*HandlerResult, error) { //nolint:gocyclo,lll,maintidx
	metricsStarts := time.Now()
	defer metrics.LogRequest(operation, metricsStarts)

	result := NewHandlerResult()

	if err := checkForMakeOperation(operation, r); err != nil {
		return result, errors.Wrap(err, "check make operation")
	}

	owner := r.Header[config.HeaderOwner]

	if len(owner) > 0 {
		log.Infof("user %s request %s", owner[0], operation)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read request body")
	}
	defer r.Body.Close()

	environment, err := api.GetEnvironmentByID(ctx, environmentID)
	if err != nil {
		return result, errors.Wrap(err, "can not get environment")
	}

	if err := r.ParseForm(); err != nil {
		return result, errors.Wrap(err, "can parsing request")
	}

	switch strings.ToLower(operation) {
	case "pods":
		pods, err := environment.GetPodsInfo(ctx)
		if err != nil {
			return result, err
		}

		result.cached = true
		result.Result = pods
	case "services":
		services, err := environment.GetServices(ctx)
		if err != nil {
			return result, err
		}

		result.Result = services
	case "info":
		result.Result = environment
	case "containers":
		filter := r.Form.Get("filter")
		containerInAnnotation := r.Form.Get("annotation")

		containers, err := environment.GetContainers(ctx, filter, containerInAnnotation)
		if err != nil {
			return result, err
		}

		result.Result = containers.Contaners
	case "debug-info":
		container := r.Form.Get("container")
		if len(container) == 0 {
			return result, errors.Wrap(errBadFormat, noContainerSpecified)
		}

		type ContainerInfo struct {
			XdebugEnabled  bool
			PhpFpmSettings string
		}

		containerInfo := ContainerInfo{}

		xdebugInfo, err := environment.ExecContainer(ctx, container, "/kubernetes-manager/xdebugInfo")
		if err != nil {
			return result, err
		}

		if xdebugInfo.Stdout != "0" {
			containerInfo.XdebugEnabled = true
		}

		phpFpmSettings, err := environment.ExecContainer(ctx, container, "/kubernetes-manager/getPhpSettings")
		if err != nil {
			return result, err
		}

		if len(phpFpmSettings.Stdout) > 0 {
			containerInfo.PhpFpmSettings = phpFpmSettings.Stdout
		}

		result.Result = containerInfo
	case "project-info":
		projectID := r.Form.Get("projectID")
		if len(projectID) == 0 {
			return result, errors.Wrap(errBadFormat, "no projectID specified")
		}

		projectInfo, err := environment.GetGitlabProjectsInfo(ctx, projectID)
		if err != nil {
			return result, err
		}

		result.Result = projectInfo
	case "make-deploy-services":
		type DeployServices struct {
			Services  string
			Operation api.GitlabPipelineOperation
		}

		deployServices := DeployServices{}

		err = json.Unmarshal(body, &deployServices)
		if err != nil {
			return result, err
		}

		if err := environment.CreateGitlabPipelinesByServices(ctx, "", deployServices.Services, deployServices.Operation); err != nil { //nolint:lll
			return result, err
		}

		result.Result = clickRefreshButton
	case "make-save-namespace-name":
		type SaveNamespaceName struct {
			Name string
		}

		saveNamespaceName := SaveNamespaceName{}

		err = json.Unmarshal(body, &saveNamespaceName)
		if err != nil {
			return result, err
		}

		if len(saveNamespaceName.Name) == 0 {
			return result, errors.Wrap(errBadFormat, "no namespace name specified")
		}

		annotations := environment.NamespaceAnnotations
		if annotations == nil {
			annotations = make(map[string]string)
		}

		annotations[config.LabelEnvironmentName] = saveNamespaceName.Name

		err := environment.SaveNamespaceMeta(ctx, annotations, environment.NamespaceLabels)
		if err != nil {
			return result, err
		}

		result.Result = "ok"
	case "make-save-user-like":
		labels := environment.NamespaceLabels
		if labels == nil {
			labels = make(map[string]string)
		}

		if len(owner) == 0 {
			return result, errors.Wrap(errBadFormat, "no owner specified")
		}

		userLabel := fmt.Sprintf("%s-%s", config.LabelUserLiked, owner[0])

		hasUserLike := false
		if labels[userLabel] == config.TrueValue {
			hasUserLike = true
		}

		labels[userLabel] = strconv.FormatBool(!hasUserLike)

		err := environment.SaveNamespaceMeta(ctx, environment.NamespaceAnnotations, labels)
		if err != nil {
			return result, err
		}

		result.Result = "ok"
	case "kubeconfig":
		kubeconfig, err := environment.GetKubeconfig(ctx)
		if err != nil {
			return result, err
		}

		kubeconfigFile, err := kubeconfig.GetRawFileContent()
		if err != nil {
			return result, err
		}

		contentDisposition := fmt.Sprintf("attachment; filename=\"kubeconfig-%s-%s\"",
			environment.Cluster,
			environment.Namespace,
		)

		result.headers["Content-Disposition"] = contentDisposition
		result.output = HandlerResultOutputRAW
		result.Result = kubeconfigFile
	case "make-pause":
		go func() {
			ctx, cancel := context.WithTimeout(backgroudContext, scaleMaxTime)
			defer cancel()

			if err := environment.ScaleALL(ctx, 0); err != nil {
				log.WithError(err).Error()
			}
		}()

		result.Result = fmt.Sprintf("All pods in namespace %s will be paused next %s", environment.Namespace, scaleMaxTime)
	case "make-start":
		go func() {
			ctx, cancel := context.WithTimeout(backgroudContext, scaleMaxTime)
			defer cancel()

			if err := environment.ScaleALL(ctx, 1); err != nil {
				log.WithError(err).Error()
			}
		}()

		result.Result = fmt.Sprintf("All pods in namespace %s will be started next %s", environment.Namespace, scaleMaxTime)
	case "make-delete":
		deleteResult := environment.DeleteALL(ctx)

		if deleteResult.HasErrors {
			return result, errors.New(deleteResult.JSON())
		}

		result.Result = fmt.Sprintf("Namespace %s deleted", environment.Namespace)
	case "make-scaledown-delay":
		type ScaledownDelay struct {
			Delay string
		}

		scaledownDelay := ScaledownDelay{}

		err := json.Unmarshal(body, &scaledownDelay)
		if err != nil {
			return result, err
		}

		durationTime, err := time.ParseDuration(scaledownDelay.Delay)
		if err != nil {
			return result, err
		}

		err = environment.ScaleDownDelay(ctx, durationTime)
		if err != nil {
			return result, err
		}

		result.Result = fmt.Sprintf("Delayed scaleDown on next %s", scaledownDelay.Delay)
	case "make-disable-hpa":
		err := environment.DisableHPA(ctx)
		if err != nil {
			return result, err
		}

		result.Result = fmt.Sprintf("Disabled HPA in namespace %s", environment.Namespace)
	case "make-disable-mtls":
		err := environment.DisableMTLS(ctx)
		if err != nil {
			return result, err
		}

		result.Result = fmt.Sprintf("Disabled mTLS in namespace %s", environment.Namespace)
	case "git-sync":
		container := r.Form.Get("container")
		if len(container) == 0 {
			return result, errors.Wrap(errBadFormat, noContainerSpecified)
		}

		type ContainerInfoResult struct {
			GitSyncEnabled bool
			PublicKey      string
			GitOrigin      string
			GitBranch      string
		}

		containerInfoResult := ContainerInfoResult{}

		containerInfo, err := environment.GetContainerInfo(ctx, container)
		if err != nil {
			return result, err
		}

		if containerInfo.PodAnnotations == nil {
			containerInfo.PodAnnotations = make(map[string]string)
		}

		containerInfoResult.GitOrigin = containerInfo.PodAnnotations[config.LabelGitSyncOrigin]
		containerInfoResult.GitBranch = containerInfo.PodAnnotations[config.LabelGitSyncBranch]

		gitSyncResult, err := environment.ExecContainer(ctx, container, "/kubernetes-manager/getGitBranch")
		if err != nil {
			return result, err
		}

		if len(gitSyncResult.Stderr) > 0 {
			return result, errors.New(gitSyncResult.Stderr)
		}

		getGitPubKey, err := environment.ExecContainer(ctx, container, "/kubernetes-manager/getGitPubKey")
		if err != nil {
			return result, err
		}

		if len(getGitPubKey.Stderr) > 0 {
			return result, errors.New(getGitPubKey.Stderr)
		}

		containerInfoResult.PublicKey = getGitPubKey.Stdout

		if len(containerInfoResult.PublicKey) > 0 {
			containerInfoResult.GitSyncEnabled = true
		}

		result.Result = containerInfoResult
	case "make-git-sync-init":
		type GitSyncInit struct {
			Container string
			GitOrigin string
			GitBranch string
		}

		gitSyncInit := GitSyncInit{}

		err = json.Unmarshal(body, &gitSyncInit)
		if err != nil {
			return result, err
		}

		if len(gitSyncInit.Container) == 0 {
			return result, errors.Wrap(errBadFormat, noContainerSpecified)
		}

		if len(gitSyncInit.GitOrigin) == 0 {
			return result, errors.Wrap(errBadFormat, "no git origin specified")
		}

		if len(gitSyncInit.GitBranch) == 0 {
			return result, errors.Wrap(errBadFormat, "no git branch specified")
		}

		enableGitCommand := fmt.Sprintf("/kubernetes-manager/enableGit %s %s", gitSyncInit.GitOrigin, gitSyncInit.GitBranch)

		enableGit, err := environment.ExecContainer(ctx, gitSyncInit.Container, enableGitCommand)
		if err != nil {
			return result, err
		}

		result.Result = enableGit.Stdout + enableGit.Stderr
	case "make-delete-container":
		type DeletePod struct {
			Container string
		}

		deletePod := DeletePod{}

		err = json.Unmarshal(body, &deletePod)
		if err != nil {
			return result, err
		}

		if len(deletePod.Container) == 0 {
			return result, errors.Wrap(errBadFormat, noContainerSpecified)
		}

		containerInfo, err := types.NewContainerInfo(deletePod.Container)
		if err != nil {
			return result, err
		}

		err = environment.DeletePod(ctx, containerInfo.PodName)
		if err != nil {
			return result, err
		}

		result.Result = fmt.Sprintf("Pod %s deleted", containerInfo.PodName)
	case "make-delete-pod":
		type DeletePod struct {
			PodName string
		}

		deletePod := DeletePod{}

		err = json.Unmarshal(body, &deletePod)
		if err != nil {
			return result, err
		}

		if len(deletePod.PodName) == 0 {
			return result, errors.Wrap(errBadFormat, "no pod specified")
		}

		err = environment.DeletePod(ctx, deletePod.PodName)
		if err != nil {
			return result, err
		}

		result.Result = fmt.Sprintf("Pod %s deleted", deletePod.PodName)

	case "make-git-sync-fetch":
		type GitSyncFetch struct {
			Container string
		}

		gitSyncFetch := GitSyncFetch{}

		err = json.Unmarshal(body, &gitSyncFetch)
		if err != nil {
			return result, err
		}

		if len(gitSyncFetch.Container) == 0 {
			return result, errors.Wrap(errBadFormat, noContainerSpecified)
		}

		gitFetch, err := environment.ExecContainer(ctx, gitSyncFetch.Container, "/kubernetes-manager/gitFetch")
		if err != nil {
			return result, err
		}

		result.Result = gitFetch.Stdout + " " + gitFetch.Stderr
	case "make-debug-xdebug-init":
		type DebugXdebugInit struct {
			Container string
		}

		debugXdebugInit := DebugXdebugInit{}

		err = json.Unmarshal(body, &debugXdebugInit)
		if err != nil {
			return result, err
		}

		if len(debugXdebugInit.Container) == 0 {
			return result, errors.Wrap(errBadFormat, noContainerSpecified)
		}

		debugXdebug, err := environment.ExecContainer(ctx, debugXdebugInit.Container, "/kubernetes-manager/enableXdebug")
		if err != nil {
			return result, err
		}

		result.Result = debugXdebug.Stdout + " " + debugXdebug.Stderr
	case "make-debug-save-config":
		type DebugSaveConfig struct {
			Container      string
			PhpFpmSettings string
		}

		debugSaveConfig := DebugSaveConfig{}

		err = json.Unmarshal(body, &debugSaveConfig)
		if err != nil {
			return result, err
		}

		if len(debugSaveConfig.Container) == 0 {
			return result, errors.Wrap(errBadFormat, noContainerSpecified)
		}

		if len(debugSaveConfig.PhpFpmSettings) == 0 {
			return result, errors.Wrap(errBadFormat, "no config specified")
		}

		base64Config := b64.StdEncoding.EncodeToString([]byte(debugSaveConfig.PhpFpmSettings))

		cmd := fmt.Sprintf("/kubernetes-manager/setPhpSettings %s", base64Config)

		debugSaveConfigExec, err := environment.ExecContainer(ctx, debugSaveConfig.Container, cmd)
		if err != nil {
			return result, err
		}

		result.Result = debugSaveConfigExec.Stdout + " " + debugSaveConfigExec.Stderr
	case "make-delete-service":
		type DeleteService struct {
			ProjectID string
			Ref       string
		}

		deleteService := DeleteService{}

		err = json.Unmarshal(body, &deleteService)
		if err != nil {
			return result, err
		}

		if len(deleteService.ProjectID) == 0 {
			return result, errors.Wrap(errBadFormat, "no project specified")
		}

		if len(deleteService.Ref) == 0 {
			return result, errors.Wrap(errBadFormat, "no ref specified")
		}

		// pipeline will be created with this environment variable:
		// DELETE=true
		// NAMESPACE=<environment.Namespace>
		// CLUSTER=<environment.Cluster>
		//
		// pipeline if succeeded, must delete namespace annotation:
		// kubectl annotate namespace $NAMESPACE kubernetes-manager/project-${CI_PROJECT_ID}-

		_, err := environment.CreateGitlabPipeline(
			ctx,
			deleteService.ProjectID,
			deleteService.Ref,
			api.GitlabPipelineOperationDelete,
		)
		if err != nil {
			return result, err
		}

		result.Result = clickRefreshButton
	case "make-git-sync-clear-cache":
		type GitSyncClearCache struct {
			Container string
		}

		gitSyncClearCache := GitSyncClearCache{}

		err = json.Unmarshal(body, &gitSyncClearCache)
		if err != nil {
			return result, err
		}

		if len(gitSyncClearCache.Container) == 0 {
			return result, errors.Wrap(errBadFormat, noContainerSpecified)
		}

		clearCache, err := environment.ExecContainer(ctx, gitSyncClearCache.Container, "/kubernetes-manager/clearCache")
		if err != nil {
			return result, err
		}

		result.Result = clearCache.Stdout + " " + clearCache.Stderr
	case "make-snapshot":
		if len(config.Get().Snapshots.ProjectID) == 0 {
			return result, errors.Wrap(errBadFormat, "no projectID for snapshoting specified")
		}

		if len(config.Get().Snapshots.Ref) == 0 {
			return result, errors.Wrap(errBadFormat, "no ref for snapshoting specified")
		}

		url, err := environment.CreateGitlabPipeline(
			ctx,
			config.Get().Snapshots.ProjectID,
			config.Get().Snapshots.Ref,
			api.GitlabPipelineOperationSnapshot,
		)
		if err != nil {
			return result, err
		}

		result.Result = fmt.Sprintf("Pipeline created %s", url)
	case "autotests":
		size := 10

		if userSize := r.Form.Get("size"); len(userSize) > 0 {
			size, err = strconv.Atoi(userSize)
			if err != nil {
				return result, errors.Wrap(err, "bad size")
			}
		}

		autotestsResults, err := autotests.GetAutotestDetails(ctx, environment, size)
		if err != nil {
			return result, err
		}

		result.Result = autotestsResults
	case "make-start-autotest":
		type StartAutotest struct {
			Test string
			User string
		}

		startAutotest := StartAutotest{}

		err = json.Unmarshal(body, &startAutotest)
		if err != nil {
			return result, err
		}

		err = autotests.StartAutotest(ctx, environment, startAutotest.Test, startAutotest.User)
		if err != nil {
			return result, err
		}

		result.Result = "Autotest started. Click Refresh button to see status."
	default:
		return result, errors.Wrap(errNoComandFound, operation)
	}

	return result, nil
}

func checkForMakeOperation(operation string, r *http.Request) error {
	if !strings.HasPrefix(operation, "make-") {
		return nil
	}

	if r.Method != http.MethodPost {
		return errMustBePOST
	}

	if _, ok := r.Header[config.HeaderOwner]; !ok {
		return errMustHaveOwner
	}

	return nil
}
