KUBECONFIG=$(HOME)/.kube/kubernetes-manager-kubeconfig
test-namespace=test-kubernetes-manager

build:
	goreleaser build --rm-dist --snapshot --skip-validate
	mv ./dist/kubernetes-manager_linux_amd64/kubernetes-manager ./kubernetes-manager
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
	go fmt ./cmd/... ./pkg/...
	go vet ./cmd/... ./pkg/...
	./scripts/test-pkg.sh
	go mod tidy
	golangci-lint run -v
	cd front && yarn lint
testIntegration:
	cp ${KUBECONFIG} ./integration-tests/testdata/kubeconfig
	kubectl delete ns ${test-namespace} || true
	kubectl create ns ${test-namespace}
	kubectl -n ${test-namespace} apply -f ./integration-tests/kubernetes
	kubectl -n ${test-namespace} wait --for=condition=available deployment --all --timeout=600s

	GOFLAGS="-count=1" POD_NAMESPACE=${test-namespace} CONFIG=testdata/config_test.yaml go test -race ./integration-tests
coverage:
	go tool cover -html=coverage.out
testChart:
	helm lint --strict ./charts/kubernetes-manager --set cleanOldTags.enabled=true
	helm lint --strict ./integration-tests/chart
	helm template ./charts/kubernetes-manager --set cleanOldTags.enabled=true | kubectl apply --dry-run=client --validate=true -f -
	helm template ./integration-tests/chart | kubectl apply --dry-run=client --validate=true -f -
install:
	helm upgrade kubernetes-manager --install --create-namespace -n kubernetes-manager ./charts/kubernetes-manager --set registry.image=paskalmaksim/kubernetes-manager:dev --set service.type=LoadBalancer
	helm upgrade kubernetes-manager-test --install --create-namespace -n kubernetes-manager-test ./integration-tests/chart
clean:
	helm uninstall kubernetes-manager -n kubernetes-manager || true
	helm uninstall kubernetes-manager -n kubernetes-manager-test || true
	kubectl delete namespace kubernetes-manager || true
	kubectl delete namespace kubernetes-manager-test || true
	kubectl delete ns ${test-namespace} || true
upgrade:
	go get -v -u k8s.io/api@v0.20.9 || true
	go get -v -u k8s.io/apimachinery@v0.20.9
	go get -v -u k8s.io/client-go@v0.20.9
	go mod tidy
	cd front && yarn update-latest
run:
	cp ${KUBECONFIG} ./kubeconfig
	POD_NAMESPACE=kubernetes-manager go run --race ./cmd/main --config=config.yaml --log.level=DEBUG $(args)
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