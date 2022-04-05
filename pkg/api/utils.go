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
	"strings"

	"k8s.io/client-go/kubernetes"
)

const (
	namespaceArrayItemCount = 2
	namespaceArraySplitter  = ":"
)

func getCluster(ns string) string {
	return strings.Split(ns, namespaceArraySplitter)[0]
}

func getNamespace(ns string) string {
	return strings.Split(ns, namespaceArraySplitter)[1]
}

func getClientset(ns string) (*kubernetes.Clientset, error) {
	namespace := strings.Split(ns, ":")
	if len(namespace) != namespaceArrayItemCount {
		return nil, errNamespaceIncorrect
	}

	clientset := clientsetCluster[getCluster(ns)]
	if clientset == nil {
		return nil, errNoCluster
	}

	return clientset, nil
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}

	return false
}
