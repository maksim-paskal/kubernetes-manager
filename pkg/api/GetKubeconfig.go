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
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"text/template"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetClusterKubeconfigResult struct {
	Endpoint    string
	CACrt       string
	CACrtBase64 string
	Token       string
}

func (r *GetClusterKubeconfigResult) GetRawFileContent() ([]byte, error) {
	kubeConfig := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: "{{ .CACrtBase64 }}"
    server: "{{ .Endpoint }}"
  name: kubernetes-manager
contexts:
- context:
    cluster: kubernetes-manager
    user: kubernetes-manager
  name: kubernetes-manager
current-context: kubernetes-manager
kind: Config
preferences: {}
users:
- name: kubernetes-manager
  user:
    token: "{{ .Token }}"`

	var out bytes.Buffer

	tmpl, err := template.New("kubeconfig").Parse(kubeConfig)
	if err != nil {
		return nil, err
	}

	err = tmpl.Execute(&out, r)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func (e *Environment) GetKubeconfig() (*GetClusterKubeconfigResult, error) {
	temporaryToken, err := e.createTemporaryToken()
	if err != nil {
		return nil, err
	}

	clusterEndpoint := "https://127.0.0.1:6443"

	for _, endpoint := range config.Get().KubernetesEndpoints {
		if endpoint.Name == e.Cluster {
			clusterEndpoint = endpoint.KubeConfigServer

			break
		}
	}

	result := GetClusterKubeconfigResult{
		Endpoint:    clusterEndpoint,
		CACrt:       string(temporaryToken.Data["ca.crt"]),
		CACrtBase64: b64.StdEncoding.EncodeToString(temporaryToken.Data["ca.crt"]),
		Token:       string(temporaryToken.Data["token"]),
	}

	return &result, nil
}

// remove old tokens with
// kubectl delete sa,role,rolebinding -A -lkubernetes-manager=true.
func (e *Environment) createTemporaryToken() (*corev1.Secret, error) {
	if e.IsSystemNamespace() {
		return nil, errors.New("cannot create temporary token in system namespace")
	}

	tokenName := fmt.Sprintf("kubernetes-manager-%s", utils.RandomString(config.TemporaryTokenRandLength))

	labels := map[string]string{
		"kubernetes-manager": "true",
	}

	serviceAccount := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tokenName,
			Labels: labels,
		},
	}

	sa, err := e.clientset.CoreV1().ServiceAccounts(e.Namespace).Create(Ctx, &serviceAccount, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error creating service account")
	}

	serviceAccountRole := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tokenName,
			Labels: labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		},
	}

	role, err := e.clientset.RbacV1().Roles(e.Namespace).Create(Ctx, &serviceAccountRole, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error creating role")
	}

	serviceAccountRoleBinding := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tokenName,
			Labels: labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     role.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	_, err = e.clientset.RbacV1().RoleBindings(e.Namespace).Create(Ctx, &serviceAccountRoleBinding, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error creating role binding")
	}

	createdUser, err := e.clientset.CoreV1().ServiceAccounts(e.Namespace).Get(Ctx, sa.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error getting user")
	}

	secret, err := e.clientset.CoreV1().Secrets(e.Namespace).Get(Ctx, createdUser.Secrets[0].Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error getting user")
	}

	return secret, nil
}