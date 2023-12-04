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

	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetEnvironmentByID(ctx context.Context, id string) (*Environment, error) {
	ctx, span := telemetry.Start(ctx, "api.GetEnvironmentByID")
	defer span.End()

	idInfo, err := types.NewIDInfo(id)
	if err != nil {
		return nil, errors.Wrap(err, "can not parse id")
	}

	clientset, err := client.GetClientset(idInfo.Cluster)
	if err != nil {
		return nil, errors.Wrap(err, "can get clientset")
	}

	e := Environment{
		ID:        id,
		Cluster:   idInfo.Cluster,
		Namespace: idInfo.Namespace,
		clientset: clientset,
	}

	namespace, err := e.clientset.CoreV1().Namespaces().Get(ctx, e.Namespace, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "can not get namespace")
	}

	if err = e.loadFromNamespace(ctx, *namespace); err != nil {
		return nil, errors.Wrap(err, "can not get namespace info")
	}

	return &e, nil
}
