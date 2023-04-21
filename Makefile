KUBECONFIG=$(HOME)/.kube/kubernetes-manager-kubeconfig
test-namespace=test-kubernetes-manager
tag=dev
image=paskalmaksim/kubernetes-manager:$(tag)
config=config.yaml
platform=linux/amd64

build:
	git tag -d `git tag -l "helm-chart-*"`
	git tag -d `git tag -l "kubernetes-manager-*"`
	go run github.com/goreleaser/goreleaser@latest build --clean --snapshot --skip-validate
	mv ./dist/kubernetes-manager_linux_amd64_v1/kubernetes-manager-amd64 .
	mv ./dist/kubernetes-manager_linux_arm64/kubernetes-manager-arm64 .
	docker buildx build --platform=$(platform) --pull --push --build-arg=APPVERSION=`git rev-parse --short HEAD` . -t $(image)
promote-to-beta:
	make build platform=linux/amd64,linux/arm64 tag=beta
security-scan:
	go run github.com/aquasecurity/trivy/cmd/trivy@latest fs --ignore-unfixed .
security-check:
	# https://github.com/aquasecurity/trivy
	go run github.com/aquasecurity/trivy/cmd/trivy@latest --ignore-unfixed $(image)
test:
	./scripts/validate-license.sh
	go fmt ./...
	go vet ./...
	./scripts/test-pkg.sh
	go mod tidy
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -v
	cd front && yarn lint

.PHONY: e2e
e2e:
	cp ${KUBECONFIG} ./e2e/testdata/kubeconfig
	kubectl delete ns ${test-namespace} || true
	kubectl create ns ${test-namespace}
	kubectl label namespace ${test-namespace} kubernetes-manager=true
	kubectl label namespace ${test-namespace} test-kubernetes-manager=true
	kubectl -n ${test-namespace} apply -f ./e2e/kubernetes
	kubectl -n ${test-namespace} wait --for=condition=available deployment --all --timeout=600s

	GOFLAGS="-count=1" POD_NAMESPACE=${test-namespace} CONFIG=testdata/config_test.yaml go test -race ./e2e
coverage:
	go tool cover -html=coverage.out
testChart:
	ct lint --charts ./charts/kubernetes-manager
	helm template ./charts/kubernetes-manager | kubectl apply --dry-run=client --validate=true -f -
install:
	helm upgrade kubernetes-manager --install --create-namespace -n kubernetes-manager ./charts/kubernetes-manager --set registry.image=$(image) --set service.type=LoadBalancer
	helm upgrade kubernetes-manager-test --install --create-namespace -n kubernetes-manager-test ./integration-tests/chart
clean:
	helm uninstall kubernetes-manager -n kubernetes-manager || true
	helm uninstall kubernetes-manager -n kubernetes-manager-test || true
	kubectl delete namespace kubernetes-manager || true
	kubectl delete namespace kubernetes-manager-test || true
	kubectl delete ns ${test-namespace} || true
upgrade:
	go get -v -u k8s.io/api@v0.23.9 || true
	go get -v -u k8s.io/apimachinery@v0.23.9
	go get -v -u k8s.io/client-go@v0.23.9
	go mod tidy
	cd front && yarn update-latest
run:
	cp ${KUBECONFIG} ./kubeconfig
	POD_NAME=kubernetes-manager POD_NAMESPACE=kubernetes-manager go run --race ./cmd/main -batch.enabled=false --config=$(config) --log.level=DEBUG $(args) --web.listen="127.0.0.1:9000"
heap:
	go tool pprof -http=127.0.0.1:8080 http://localhost:9000/debug/pprof/heap
allocs:
	go tool pprof -http=127.0.0.1:8080 http://localhost:9000/debug/pprof/heap
chart-index:
	rm -rf .cr-index
	mkdir .cr-index
	cr index \
	--owner maksim-paskal \
	--git-repo kubernetes-manager \
	--release-name-template "helm-chart-{{ .Version }}" \
	--charts-repo https://maksim-paskal.github.io/kubernetes-manager \
	--push \
	--token $(CR_TOKEN)
chart-upload:
	rm -rf .cr-release-packages
	cr package ./charts/kubernetes-manager
	cr upload \
	--owner maksim-paskal \
	--git-repo kubernetes-manager \
	--commit "`git rev-parse HEAD`" \
	--release-name-template "helm-chart-{{ .Version }}" \
	--token $(CR_TOKEN)
scan:
	@trivy image \
	-ignore-unfixed --no-progress --severity HIGH,CRITICAL \
	$(image)
	@helm template ./charts/kubernetes-manager > /tmp/kubernetes-manager.yaml
	@trivy config /tmp/kubernetes-manager.yaml
	@trivy fs -ignore-unfixed --no-progress --severity HIGH,CRITICAL front/yarn.lock
	@trivy fs -ignore-unfixed --no-progress --severity HIGH,CRITICAL go.sum