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
	"context"
	"fmt"
	"slices"

	"github.com/maksim-paskal/kubernetes-manager/pkg/cache"
	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/metrics"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetCachedGitlabProject(ctx context.Context, projectID string) (*gitlab.Project, error) {
	ctx, span := telemetry.Start(ctx, "api.GetCachedGitlabProject")
	defer span.End()

	cacheKey := "gitlab::project::" + projectID

	var cacheValue gitlab.Project

	if err := cache.Client().Get(ctx, cacheKey, &cacheValue); err == nil {
		metrics.CacheHits.WithLabelValues("GetCachedGitlabProject").Inc()

		return &cacheValue, nil
	}

	project, _, err := client.GetGitlabClient().Projects.GetProject(
		projectID,
		&gitlab.GetProjectOptions{},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, errors.Wrap(err, "can not get project")
	}

	_ = cache.Client().Set(ctx, cacheKey, project, cache.HighTTL)

	return project, nil
}

func GetCachedGitlabPipelineVariables(ctx context.Context, projectID string, pipeline int) ([]*gitlab.PipelineVariable, error) {
	ctx, span := telemetry.Start(ctx, "api.GetCachedGitlabPipelineVariables")
	defer span.End()

	cacheKey := fmt.Sprintf("gitlab::project::%s::pipeline::%d::variables", projectID, pipeline)
	cacheValue := make([]*gitlab.PipelineVariable, 0)

	if err := cache.Client().Get(ctx, cacheKey, &cacheValue); err == nil {
		metrics.CacheHits.WithLabelValues("GetCachedGitlabPipelineVariables").Inc()

		return cacheValue, nil
	}

	pipelineVars, _, err := client.GetGitlabClient().Pipelines.GetPipelineVariables(
		projectID,
		pipeline,
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project pipeline variables")
	}

	_ = cache.Client().Set(ctx, cacheKey, pipelineVars, cache.MaxTTL)

	return pipelineVars, nil
}

func GetCachedGitlabProjectsByTopic(ctx context.Context, topic string) ([]*gitlab.Project, error) {
	ctx, span := telemetry.Start(ctx, "api.GetCachedGitlabProjectsByTopic")
	defer span.End()

	cacheKey := "gitlab::projects::topic::" + topic
	cacheValue := make([]*gitlab.Project, 0)

	if err := cache.Client().Get(ctx, cacheKey, &cacheValue); err == nil {
		metrics.CacheHits.WithLabelValues("GetCachedGitlabProjectsByTopic").Inc()

		return cacheValue, nil
	}

	projects, _, err := client.GetGitlabClient().Projects.ListProjects(
		&gitlab.ListProjectsOptions{
			Topic: gitlab.Ptr(topic),
		},
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return nil, errors.Wrap(err, "can not list projects")
	}

	_ = cache.Client().Set(ctx, cacheKey, projects, cache.HighTTL)

	return projects, nil
}

type PodStatus string

const (
	// return only running pods.
	PodIsRunning PodStatus = "running"
	// return all pods except Succeeded.
	PodIsNotSucceeded PodStatus = "not_succeeded"
)

func GetCachedKubernetesPodsStatus(ctx context.Context, cluster, namespace string, status PodStatus) ([]corev1.Pod, error) {
	ctx, span := telemetry.Start(ctx, "api.GetCachedKubernetesPodsRunning")
	defer span.End()

	pods, err := GetCachedKubernetesPodsByFieldSelector(ctx, cluster, namespace, "")
	if err != nil {
		return nil, errors.Wrap(err, "can not list pods")
	}

	// filter all pods
	result := slices.DeleteFunc(pods, func(pod corev1.Pod) bool {
		switch status {
		case PodIsRunning:
			return !(pod.Status.Phase == corev1.PodRunning)
		case PodIsNotSucceeded:
			return !(pod.Status.Phase != corev1.PodSucceeded)
		}

		return false
	})

	return result, nil
}

func GetCachedKubernetesPodsByFieldSelector(ctx context.Context, cluster, namespace, selector string) ([]corev1.Pod, error) {
	ctx, span := telemetry.Start(ctx, "api.GetCachedKubernetesPodsByFieldSelector")
	defer span.End()

	cacheKey := fmt.Sprintf("kubernetes::pods::%s::%s::%s", cluster, namespace, selector)
	cacheValue := make([]corev1.Pod, 0)

	if err := cache.Client().Get(ctx, cacheKey, &cacheValue); err == nil {
		metrics.CacheHits.WithLabelValues("GetCachedKubernetesPodsByFieldSelector").Inc()

		return cacheValue, nil
	}

	clientset, err := client.GetClientset(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "can not get clientset")
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "can not list pods")
	}

	_ = cache.Client().Set(ctx, cacheKey, pods.Items, cache.LowTTL)

	return pods.Items, nil
}

func GetCachedPersistentVolumeClaims(ctx context.Context, cluster, namespace string) ([]corev1.PersistentVolumeClaim, error) {
	ctx, span := telemetry.Start(ctx, "api.GetCachedPersistentVolumeClaims")
	defer span.End()

	cacheKey := fmt.Sprintf("kubernetes::pvcs::%s::%s", cluster, namespace)
	cacheValue := make([]corev1.PersistentVolumeClaim, 0)

	if err := cache.Client().Get(ctx, cacheKey, &cacheValue); err == nil {
		metrics.CacheHits.WithLabelValues("GetCachedPersistentVolumeClaims").Inc()

		return cacheValue, nil
	}

	clientset, err := client.GetClientset(cluster)
	if err != nil {
		return nil, errors.Wrap(err, "can not get clientset")
	}

	pvc, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error list pvc")
	}

	_ = cache.Client().Set(ctx, cacheKey, pvc.Items, cache.LowTTL)

	return pvc.Items, nil
}
