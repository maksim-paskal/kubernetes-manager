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
	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	opentracing "github.com/opentracing/opentracing-go"
)

type execContainerParams struct {
	namespace     string
	labelSelector string
	podname       string
	container     string
	command       string
}

func execContainer(rootSpan opentracing.Span, params execContainerParams) (*api.ExecContainerResults, error) {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("execContainer", opentracing.ChildOf(rootSpan.Context()))

	defer span.Finish()

	return api.ExecContainer(
		params.namespace,
		params.podname,
		params.labelSelector,
		params.container,
		params.command,
	)
}
