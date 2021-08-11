KUBECONFIG=$(HOME)/.kube/example-kubeconfig

build:
	docker build --pull . -t paskalmaksim/kubernetes-manager:dev
security-scan:
	trivy fs --ignore-unfixed .
security-check:
	# https://github.com/aquasecurity/trivy
	trivy --ignore-unfixed paskalmaksim/kubernetes-manager:dev
push:
	docker push paskalmaksim/kubernetes-manager:dev
test:
	./scripts/validate-license.sh
	./scripts/test-pkg.sh
	go mod tidy
	golangci-lint run -v
	cd front && yarn lint
testChart:
	helm lint --strict ./chart/kubernetes-manager
	helm lint --strict ./chart/kubernetes-manager-test
	helm template ./chart/kubernetes-manager | kubectl apply --dry-run=client --validate=true -f -
	helm template ./chart/kubernetes-manager-test | kubectl apply --dry-run=client --validate=true -f -
install:
	helm upgrade kubernetes-manager --install --create-namespace -n kubernetes-manager ./chart/kubernetes-manager --set registry.image=paskalmaksim/kubernetes-manager:dev --set service.type=LoadBalancer
	helm upgrade kubernetes-manager-test --install --create-namespace -n kubernetes-manager-test ./chart/kubernetes-manager-test
clean:
	helm uninstall kubernetes-manager -n kubernetes-manager
	helm uninstall kubernetes-manager -n kubernetes-manager-test
	kubectl delete namespace kubernetes-manager
	kubectl delete namespace kubernetes-manager-test
build-all:
	scripts/build-all.sh
upgrade:
	go get -v -u k8s.io/api@v0.20.9 || true
	go get -v -u k8s.io/apimachinery@v0.20.9
	go get -v -u k8s.io/client-go@v0.20.9
	go mod tidy
	cd front && yarn update-latest
run:
	POD_NAMESPACE=kubernetes-manager go run --race ./cmd/main --log.level=DEBUG --kubeconfig.path=$(KUBECONFIG) $(args)