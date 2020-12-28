build:
	docker build --pull . -t paskalmaksim/kubernetes-manager:dev
test:
	./scripts/validate-license.sh
	go fmt ./cmd/main
	go mod tidy
	go test ./cmd/main
	golangci-lint run --allow-parallel-runners -v --enable-all --disable testpackage,nestif,gochecknoglobals,funlen,gocognit,exhaustivestruct --fix
	cd front && yarn lint
testChart:
	helm lint --strict ./chart/kubernetes-manager
	helm template ./chart/kubernetes-manager | kubectl apply --dry-run --validate=true -f -
build-all:
	scripts/build-all.sh
upgrade:
	go get -v -u all
	# downgrade to v0.18.14
	go get -v -u k8s.io/api@v0.18.14 || true
	go get -v -u k8s.io/apimachinery@v0.18.14
	go get -v -u k8s.io/client-go@v0.18.14
	# downgrade for k8s.io/client-go@v0.18.14
	go get -v -u github.com/googleapis/gnostic@v0.1.0
	go mod tidy
	cd front && yarn update-latest