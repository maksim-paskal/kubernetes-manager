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
	b64 "encoding/base64"
	"os"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetClusterKubeconfigResult struct {
	Endpoint    string
	CACrt       string
	CACrtBase64 string
	Token       string
}

func GetClusterKubeconfig(cluster string) (*GetClusterKubeconfigResult, error) {
	clientset := clientsetCluster[cluster]
	if clientset == nil {
		return &GetClusterKubeconfigResult{}, errNoCluster
	}

	namespace := os.Getenv("POD_NAMESPACE")

	sa, err := clientset.CoreV1().ServiceAccounts(namespace).Get(Ctx, config.Namespace, metav1.GetOptions{})
	if err != nil {
		return &GetClusterKubeconfigResult{}, errors.Wrap(err, "error getting service account")
	}

	secretName := sa.Secrets[0].Name

	secret, err := clientset.CoreV1().Secrets(namespace).Get(Ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return &GetClusterKubeconfigResult{}, errors.Wrap(err, "error getting secret")
	}

	clusterEndpoint := ""

	for _, endpoint := range config.Get().KubernetesEndpoints {
		if endpoint.Name == cluster {
			clusterEndpoint = endpoint.KubeConfigServer

			break
		}
	}

	result := GetClusterKubeconfigResult{
		Endpoint:    clusterEndpoint,
		CACrt:       string(secret.Data["ca.crt"]),
		CACrtBase64: b64.StdEncoding.EncodeToString(secret.Data["ca.crt"]),
		Token:       string(secret.Data["token"]),
	}

	return &result, nil
}
