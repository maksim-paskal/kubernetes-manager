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
	b64 "encoding/base64"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
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

func (r *GetClusterKubeconfigResult) GetRawFileContent(ctx context.Context) ([]byte, error) {
	ctx, span := telemetry.Start(ctx, "api.GetRawFileContent")
	defer span.End()

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

	result, err := utils.GetTemplatedResult(ctx, kubeConfig, r)
	if err != nil {
		return nil, errors.Wrap(err, "error getting templated string")
	}

	return result, nil
}

func (e *Environment) GetKubeconfig(ctx context.Context) (*GetClusterKubeconfigResult, error) {
	ctx, span := telemetry.Start(ctx, "api.GetKubeconfig")
	defer span.End()

	temporaryToken, err := e.createTemporaryToken(ctx)
	if err != nil {
		return nil, err
	}

	clusterEndpoint := "https://127.0.0.1:6443"

	if endpoint := config.Get().GetKubernetesEndpointByName(e.Cluster); endpoint != nil {
		clusterEndpoint = endpoint.KubeConfigServer
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
func (e *Environment) createTemporaryToken(ctx context.Context) (*corev1.Secret, error) {
	ctx, span := telemetry.Start(ctx, "api.createTemporaryToken")
	defer span.End()

	if e.IsSystemNamespace() {
		return nil, errors.New("cannot create temporary token in system namespace")
	}

	tokenName := "kubernetes-manager-" + utils.RandomString(config.TemporaryTokenRandLength)

	labels := map[string]string{
		"kubernetes-manager": "true",
	}

	serviceAccount := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tokenName,
			Labels: labels,
		},
	}

	sa, err := e.clientset.CoreV1().ServiceAccounts(e.Namespace).Create(ctx, &serviceAccount, metav1.CreateOptions{})
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

	role, err := e.clientset.RbacV1().Roles(e.Namespace).Create(ctx, &serviceAccountRole, metav1.CreateOptions{})
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

	_, err = e.clientset.RbacV1().RoleBindings(e.Namespace).Create(ctx, &serviceAccountRoleBinding, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error creating role binding")
	}

	saSecret := corev1.Secret{
		Type: corev1.SecretTypeServiceAccountToken,
		ObjectMeta: metav1.ObjectMeta{
			Name:   sa.Name,
			Labels: labels,
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": sa.Name,
			},
		},
	}

	token, err := e.clientset.CoreV1().Secrets(e.Namespace).Create(ctx, &saSecret, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error creating token")
	}

	secret, err := e.clientset.CoreV1().Secrets(e.Namespace).Get(ctx, token.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error getting secrets")
	}

	return secret, nil
}
