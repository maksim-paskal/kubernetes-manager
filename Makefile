build:
	docker build --pull . -t paskalmaksim/kubernetes-manager:dev
test:
	./scripts/validate-license.sh
	go fmt ./cmd/main
	go mod tidy
	go test ./cmd/main
	golangci-lint run --allow-parallel-runners -v --enable-all --disable testpackage,wsl,maligned,nestif,gochecknoglobals,funlen,gocognit --fix
testChart:
	helm lint --strict ./chart/kubernetes-manager
	helm template ./chart/kubernetes-manager | kubectl apply --dry-run --validate=true -f -
build-all:
	scripts/build-all.sh