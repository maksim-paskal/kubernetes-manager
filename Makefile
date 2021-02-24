build:
	docker build --pull . -t paskalmaksim/kubernetes-manager:dev
test:
	./scripts/validate-license.sh
	go fmt ./cmd/main
	go mod tidy
	go test ./cmd/main
	golangci-lint run -v
	cd front && yarn lint
testChart:
	helm lint --strict ./chart/kubernetes-manager
	helm template ./chart/kubernetes-manager | kubectl apply --dry-run --validate=true -f -
build-all:
	scripts/build-all.sh
upgrade:
	go get -v -u k8s.io/api@v0.19.8 || true
	go get -v -u k8s.io/apimachinery@v0.19.8
	go get -v -u k8s.io/client-go@v0.19.8
	go mod tidy
	cd front && yarn update-latest