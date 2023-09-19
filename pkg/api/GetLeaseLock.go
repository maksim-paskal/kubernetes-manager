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
	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

var errNoPodNamespaceOrPodName = errors.New("no pod namespace or pod name")

func GetLeaseLock(podNamespace string, podName string) (*resourcelock.LeaseLock, error) {
	if len(podNamespace) == 0 || len(podName) == 0 {
		return nil, errNoPodNamespaceOrPodName
	}

	clientset, err := client.GetInclusterClientset()
	if err != nil {
		return nil, errors.Wrap(err, "error getting clientset")
	}

	return &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      config.Namespace,
			Namespace: podNamespace,
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: podName,
		},
	}, nil
}
