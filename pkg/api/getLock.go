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
	"errors"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

const lockName = "kubernetes-manager"

var (
	errNoPodNamespaceOrPodName = errors.New("no pod namespace or pod name")
	errNoKubernetesEndpoints   = errors.New("no kubernetes endpoints")
)

func GetLeaseLock(podNamespace string, podName string) (*resourcelock.LeaseLock, error) {
	if len(podNamespace) == 0 || len(podName) == 0 {
		return nil, errNoPodNamespaceOrPodName
	}

	if len(config.Get().KubernetesEndpoints) == 0 {
		return nil, errNoKubernetesEndpoints
	}

	clusterForLeaderElection := config.Get().KubernetesEndpoints[0].Name

	return &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      lockName,
			Namespace: podNamespace,
		},
		Client: clientsetCluster[clusterForLeaderElection].CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: podName,
		},
	}, nil
}
