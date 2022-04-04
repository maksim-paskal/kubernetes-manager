# Kubernetes Manager

The manager of feature branches for production teams in kubernetes cluster

## Motivation

There are many feature branches in cluster - and it's sometime hard to detect what feature is currently running, and how to access this feature. This tool can manage you branches in kubernetes cluster

![kubernetes-manager-gui](https://raw.githubusercontent.com/maksim-paskal/artwork/master/kubernetes-manager-gui.png)

## Get Started

### Kubernetes Manager installation

```bash
helm repo add maksim-paskal-kubernetes-manager https://maksim-paskal.github.io/kubernetes-manager
helm repo update

helm upgrade kubernetes-manager \
  --install \
  --create-namespace \
  --namespace kubernetes-manager \
  maksim-paskal-kubernetes-manager/kubernetes-manager \
  --set service.type=LoadBalancer
```

you need to get your new LoadBalancer address - and open your browser `http://<LoadBalancerAddress>:9000`

### Test kubernetes-manager with example Ingress

```bash
helm upgrade kubernetes-manager-test \
  --install \
  --create-namespace \
  --namespace kubernetes-manager-test \
  ./integration-tests/chart
```

### Setup you own Ingress

you need to add annotation and label to your Ingress controller

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-cool-product
  labels:
    # required to get listed in kubernetes-manager
    kubernetes-manager: "true"
  annotations:
    # your git repo
    kubernetes-manager/git-project-origin: "git@github.com:maksim-paskal/kubernetes-manager.git"
    # git repo branch
    kubernetes-manager/git-branch: some-git-tag
    # minimum pod count in namespace for calculation that feature branch is running
    kubernetes-manager/requiredRunningPodsCount: "1"
    ################################################################################
    # not required 
    ################################################################################
    # if you using Gitlab - you need to pass gitlab project id
    kubernetes-manager/git-project-id: 1
    # docker tag of feature branch
    kubernetes-manager/registry-tag: some-docker-tag
    # default container in namespace (format: <podLabelKey>=<podLabelValue>:<containerName>)
    kubernetes-manager/default-pod: app=nginx:nginx
...
```

and give kubernetes-manager access to manage your namespace

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: kubernetes-manager-role
rules:
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["delete","patch"]
- apiGroups: [""]
  resources: ["pods","services"]
  verbs: ["get","list","delete"]
- apiGroups: ["apps"]
  resources: ["deployments","statefulsets"]
  verbs: ["list","get","update"]
- apiGroups: ["apps"]
  resources: ["deployments/scale"]
  verbs: ["patch"]
- apiGroups: [""]
  resources: ["pods/exec","pods/portforward"]
  verbs: ["create"]
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get"]
- apiGroups: ["autoscaling"]
  resources: ["horizontalpodautoscalers"]
  verbs: ["list","delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: kubernetes-manager
roleRef:
  kind: Role
  name: kubernetes-manager-role
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  name: kubernetes-manager
  # namespace where kubernetes-manager is running
  namespace: kubernetes-manager
```

## Extentions

if you using Gitlab as git repo - you can pass environment variables to kubernetes-manager to reduce resources of you kubernetes cluster, and reduce disk usage of docker registry

```bash
# gitlab api endpoint
GITLAB_URL=https://git/api/v4
# api token
GITLAB_TOKEN=some-token
```

[Delete feature branch when it was merged into main](pkg/batch/README.md)

[Clear old docker registry tags](https://github.com/maksim-paskal/gitlab-registry-cleaner)

## Development environment

### start front server

```bash
cd front
yarn install
yarn dev
```

### start backend server

```bash
make run KUBECONFIG=/path/to/kubeconfig
```

open your browser `http://127.0.0.1:3000`
