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
package client

import (
	"net/http"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/metrics"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	k8sMetrics "k8s.io/client-go/tools/metrics"
)

var (
	gitlabClient *gitlab.Client
	hcloudClient *hcloud.Client

	clientsetCluster  map[string]*kubernetes.Clientset
	restconfigCluster map[string]*rest.Config
)

var gitlabHTTPClient = &http.Client{
	Jar:       nil,
	Transport: metrics.NewInstrumenter("gitlab").InstrumentedRoundTripper(),
}

var hcloudHTTPClient = &http.Client{
	Jar:       nil,
	Transport: metrics.NewInstrumenter("hcloud").InstrumentedRoundTripper(),
}

var errNoCluster = errors.New("no cluster")

func GetGitlabClient() *gitlab.Client {
	return gitlabClient
}

func GetAllClientsets() map[string]*kubernetes.Clientset {
	return clientsetCluster
}

func GetClientset(cluster string) (*kubernetes.Clientset, error) {
	clientset := clientsetCluster[cluster]
	if clientset == nil {
		return nil, errNoCluster
	}

	return clientset, nil
}

func GetRestConfig(cluster string) (*rest.Config, error) {
	restconfig := restconfigCluster[cluster]
	if restconfig == nil {
		return nil, errNoCluster
	}

	return restconfig, nil
}

func GetInclusterClientset() (*kubernetes.Clientset, error) {
	restconfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return nil, errors.Wrap(err, "can not create connection")
	}

	return clientset, nil
}

func GetHcloudClient() *hcloud.Client {
	return hcloudClient
}

func Init() error {
	var (
		err        error
		restconfig *rest.Config
	)

	if len(*config.Get().GitlabToken) > 0 || len(*config.Get().GitlabURL) > 0 {
		gitlabClient, err = gitlab.NewClient(
			*config.Get().GitlabToken,
			gitlab.WithBaseURL(*config.Get().GitlabURL),
			gitlab.WithHTTPClient(gitlabHTTPClient),
		)
		if err != nil {
			return errors.Wrap(err, "can not connect to Gitlab")
		}
	}

	k8sMetrics.Register(k8sMetrics.RegisterOpts{
		RequestResult:  &requestResult{},
		RequestLatency: &requestLatency{},
	})

	clientsetCluster = make(map[string]*kubernetes.Clientset)
	restconfigCluster = make(map[string]*rest.Config)

	for _, kubernetesEndpoints := range config.Get().GetKubernetesEndpoints() {
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

		inst := metrics.NewInstrumenter("k8s" + kubernetesEndpoints.Name)

		// don't log all path
		inst.PathDetectionFunc = func(_ *http.Request) string {
			return ""
		}

		restconfig.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
			return inst.InstrumentedRoundTripperWithTransport(rt)
		}

		clientset, err := kubernetes.NewForConfig(restconfig)
		if err != nil {
			return errors.Wrap(err, "can not create connection")
		}

		clientsetCluster[kubernetesEndpoints.Name] = clientset
		restconfigCluster[kubernetesEndpoints.Name] = restconfig
	}

	hcloudClient = hcloud.NewClient(
		hcloud.WithToken(config.Get().RemoteServer.HetznerToken),
		hcloud.WithHTTPClient(hcloudHTTPClient),
	)

	return nil
}
