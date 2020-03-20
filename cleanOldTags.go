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
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* TODO: not deleted tags */
func cleanOldTags(rootSpan opentracing.Span) {
	var tracer = opentracing.GlobalTracer()
	span := tracer.StartSpan("cleanOldTagsBy", opentracing.ChildOf(rootSpan.Context()))
	defer span.Finish()

	projectIDs := []string{}

	opt := metav1.ListOptions{
		LabelSelector: *appConfig.ingressFilter,
	}

	ingresss, _ := clientset.ExtensionsV1beta1().Ingresses("").List(opt)
	for _, ingress := range ingresss.Items {
		projectID := ingress.Annotations["kubernetes-manager/git-project-id"]

		if !stringInSlice(projectID, projectIDs) {
			projectIDs = append(projectIDs, ingress.Annotations["kubernetes-manager/git-project-id"])
		}
	}

	for _, projectID := range projectIDs {
		cleanOldTagsByProject(rootSpan, projectID)
	}
}
func cleanOldTagsByProject(rootSpan opentracing.Span, projectID string) {
	var tracer = opentracing.GlobalTracer()
	span := tracer.StartSpan("cleanOldTagsByProject", opentracing.ChildOf(rootSpan.Context()))
	defer span.Finish()

	/* ADD TAGS NOT DELETE */
	nonDelete := []string{}
	opt := metav1.ListOptions{
		LabelSelector: *appConfig.ingressFilter,
	}

	ingresss, _ := clientset.ExtensionsV1beta1().Ingresses("").List(opt)
	for _, ingress := range ingresss.Items {
		if ingress.Annotations["kubernetes-manager/git-project-id"] == projectID {
			nonDelete = append(nonDelete, ingress.Annotations["kubernetes-manager/registry-tag"])
		}
	}

	log.Infof("tags to delete=%s", nonDelete)

	git := gitlab.NewClient(nil, *appConfig.gitlabToken)
	err := git.SetBaseURL(*appConfig.gitlabURL)
	if err != nil {
		log.Panic(err)
	}
	gitRepos, _, err := git.ContainerRegistry.ListRegistryRepositories(projectID, nil)

	if err != nil {
		log.Panic(err)
	}

	for _, gitRepo := range gitRepos {
		gitRepoTags, _, err := git.ContainerRegistry.ListRegistryRepositoryTags(projectID, gitRepo.ID, nil)

		if err != nil {
			log.Panic(err)
		}

		for _, gitRepoTag := range gitRepoTags {
			if !stringInSlice(gitRepoTag.Name, nonDelete) {

				log.Infof("Delete project=%s repo=%d tag=%s", projectID, gitRepo.ID, gitRepoTag.Name)

				_, err := git.ContainerRegistry.DeleteRegistryRepositoryTag(projectID, gitRepo.ID, gitRepoTag.Name)

				if err != nil {
					log.Error(err)
					span.LogKV("warning", err)
				}
			}
		}
	}
}
