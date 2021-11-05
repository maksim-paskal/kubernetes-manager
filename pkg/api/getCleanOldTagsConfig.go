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
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetCleanOldTagsConfig() ([]string, error) {
	result := make([]string, 0)

	opt := metav1.ListOptions{
		LabelSelector: "app=cleanoldtags",
	}

	for _, clientset := range clientsetCluster {
		cms, err := clientset.CoreV1().ConfigMaps(os.Getenv("POD_NAMESPACE")).List(Ctx, opt)
		if err != nil {
			return nil, err
		}

		for _, cm := range cms.Items {
			podNamespace := os.Getenv("POD_NAMESPACE")

			cleanoldtags, err := clientset.
				CoreV1().
				ConfigMaps(podNamespace).
				Get(Ctx, cm.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}

			result = append(result, cleanoldtags.Data["exceptions"])
		}
	}

	return result, nil
}
