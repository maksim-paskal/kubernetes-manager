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
	"net/url"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func batch(rootSpan opentracing.Span) {
	var tracer = opentracing.GlobalTracer()
	span := tracer.StartSpan("batch", opentracing.ChildOf(rootSpan.Context()))
	defer span.Finish()

	git := gitlab.NewClient(nil, *appConfig.gitlabToken)
	err := git.SetBaseURL(*appConfig.gitlabURL)
	if err != nil {
		log.Panic(err)
	}
	opt := metav1.ListOptions{
		LabelSelector: *appConfig.ingressFilter,
	}

	ingresss, _ := clientset.ExtensionsV1beta1().Ingresses("").List(opt)

	for _, ingress := range ingresss.Items {
		gitBranch := ingress.Annotations["kubernetes-manager/git-branch"]
		gitProjectID := ingress.Annotations["kubernetes-manager/git-project-id"]

		_, _, err := git.Branches.GetBranch(gitProjectID, gitBranch)
		if err != nil {
			if strings.Contains(err.Error(), "404 Branch Not Found") {
				span.LogKV("delete branch", gitBranch)

				ch1 := make(chan httpResponse)

				q := make(url.Values)

				q.Add("namespace", ingress.Namespace)

				for k, v := range ingress.Annotations {
					if strings.HasPrefix(k, "kubernetes-manager") {
						q.Add(k[19:], v)
					}
				}

				go makeAPICall(span, "/api/deleteALL", q, ch1)

				span.LogKV("result", <-ch1)
			}
		}
	}
}
