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

## Extentions

if you using Gitlab as git repo - you can pass environment variables to kubernetes-manager to reduce resources of you kubernetes cluster, and reduce disk usage of docker registry

```bash
# gitlab api endpoint
GITLAB_URL=https://git/api/v4
# api token
GITLAB_TOKEN=some-token
```

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
