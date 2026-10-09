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
	"github.com/maksim-paskal/sluglify"
	"github.com/pkg/errors"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const kubeRootCAConfigMapName = "kube-root-ca.crt"

const maxTokenOwnerLength = 20

type apiGroupResources struct {
	group     string
	resources []string
}

// resources visible through a read-only token in a system namespace, grouped by the API
// group they belong to. RBAC matches APIGroups and Resources independently rather than
// as paired tuples, so a single APIGroups:["*"] rule listing these names would also grant
// access to any CRD that happens to share a plural name (e.g. a jobs.somevendor.io CRD) -
// scoping each group explicitly avoids that.
// deliberately excludes secrets, roles, rolebindings, serviceaccounts and configmaps
// (serviceaccounts can reveal cloud IAM role bindings via annotations, e.g. AWS IRSA;
// configmaps can hold sensitive data, e.g. EKS's aws-auth).
var systemNamespaceReadOnlyResources = []apiGroupResources{
	{
		group: "",
		resources: []string{
			"pods", "pods/log", "pods/status",
			"services", "endpoints",
			"events",
			"persistentvolumeclaims",
			"replicationcontrollers",
			"resourcequotas",
			"limitranges",
		},
	},
	{
		group: "apps",
		resources: []string{
			"deployments", "deployments/scale", "deployments/status",
			"replicasets", "replicasets/status",
			"statefulsets", "statefulsets/status",
			"daemonsets",
		},
	},
	{
		group: "batch",
		resources: []string{
			"jobs", "jobs/status",
			"cronjobs", "cronjobs/status",
		},
	},
	{
		group:     "networking.k8s.io",
		resources: []string{"ingresses", "ingresses/status", "networkpolicies"},
	},
	{
		group:     "autoscaling",
		resources: []string{"horizontalpodautoscalers"},
	},
}

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

	// fetched before provisioning any RBAC objects so a failure here never leaves an
	// orphaned ServiceAccount/Role/RoleBinding behind.
	caCrt, err := e.getClusterCACert(ctx)
	if err != nil {
		return nil, err
	}

	tokenRequest, err := e.createTemporaryToken(ctx)
	if err != nil {
		return nil, err
	}

	clusterEndpoint := "https://127.0.0.1:6443"

	if endpoint := config.Get().GetKubernetesEndpointByName(e.Cluster); endpoint != nil {
		clusterEndpoint = endpoint.KubeConfigServer
	}

	result := GetClusterKubeconfigResult{
		Endpoint:    clusterEndpoint,
		CACrt:       string(caCrt),
		CACrtBase64: b64.StdEncoding.EncodeToString(caCrt),
		Token:       tokenRequest.Status.Token,
	}

	return &result, nil
}

// every namespace has a kube-root-ca.crt configmap (published by the RootCAConfigMap
// controller, enabled by default since Kubernetes 1.20) containing the cluster CA cert.
func (e *Environment) getClusterCACert(ctx context.Context) ([]byte, error) {
	ctx, span := telemetry.Start(ctx, "api.getClusterCACert")
	defer span.End()

	configMap, err := e.clientset.CoreV1().ConfigMaps(e.Namespace).Get(ctx, kubeRootCAConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error getting cluster CA certificate")
	}

	return []byte(configMap.Data["ca.crt"]), nil
}

// remove old tokens with
// kubectl delete sa,role,rolebinding -A -lkubernetes-manager=true.
func (e *Environment) createTemporaryToken(ctx context.Context) (*authenticationv1.TokenRequest, error) {
	ctx, span := telemetry.Start(ctx, "api.createTemporaryToken")
	defer span.End()

	randSuffix := utils.RandomString(config.TemporaryTokenRandLength)
	tokenName := "kubernetes-manager-" + randSuffix

	// embed the requesting user in the name so that a Forbidden error from this
	// token (e.g. `system:serviceaccount:<ns>:<name>`) names who requested it.
	if owner := sluglify.GetSlugString(e.GetUser(ctx), maxTokenOwnerLength, ""); len(owner) > 0 {
		tokenName = "kubernetes-manager-" + owner + "-" + randSuffix
	}

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

	var rules []rbacv1.PolicyRule

	// system namespaces only get read-only access, and never to secrets.
	if e.IsSystemNamespace() {
		rules = make([]rbacv1.PolicyRule, 0, len(systemNamespaceReadOnlyResources))

		for _, apiGroup := range systemNamespaceReadOnlyResources {
			rules = append(rules, rbacv1.PolicyRule{
				APIGroups: []string{apiGroup.group},
				Resources: apiGroup.resources,
				Verbs:     []string{"get", "list", "watch"},
			})
		}
	} else {
		rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		}
	}

	serviceAccountRole := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tokenName,
			Labels: labels,
		},
		Rules: rules,
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

	const secondsInHour = 3600

	// bake a real exp claim into the JWT so it is invalid after TemporaryTokenDurationHours,
	// independent of the external cleanup job in DeleteTemporaryTokens.
	expirationSeconds := int64(config.TemporaryTokenDurationHours * secondsInHour)

	tokenRequest, err := e.clientset.CoreV1().ServiceAccounts(e.Namespace).CreateToken(ctx, sa.Name, &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			ExpirationSeconds: &expirationSeconds,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error creating token")
	}

	return tokenRequest, nil
}
