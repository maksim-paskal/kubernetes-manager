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
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	k8sMetrics "k8s.io/client-go/tools/metrics"
)

func Init() error {
	var (
		err        error
		restconfig *rest.Config
	)

	k8sMetrics.Register(k8sMetrics.RegisterOpts{
		RequestResult:  &requestResult{},
		RequestLatency: &requestLatency{},
	})

	clientsetCluster = make(map[string]*kubernetes.Clientset)
	restconfigCluster = make(map[string]*rest.Config)

	for _, kubernetesEndpoints := range config.Get().KubernetesEndpoints {
		if len(kubernetesEndpoints.KubeConfigPath) > 0 {
			restconfig, err = clientcmd.BuildConfigFromFlags("", kubernetesEndpoints.KubeConfigPath)
			if err != nil {
				return err
			}
		} else {
			log.Info("No kubeconfig file use incluster")
			restconfig, err = rest.InClusterConfig()
			if err != nil {
				return err
			}
		}

		clientset, err := kubernetes.NewForConfig(restconfig)
		if err != nil {
			return errors.Wrap(err, "can not create connection")
		}

		clientsetCluster[kubernetesEndpoints.Name] = clientset
		restconfigCluster[kubernetesEndpoints.Name] = restconfig
	}

	return nil
}
