build:
	docker build --pull . -t paskalmaksim/kubernetes-manager:dev
test:
	go test ./cmd/main
	golangci-lint run
	scripts/validate-license.sh
testChart:
	helm lint --strict ./chart/kubernetes-manager
	helm template ./chart/kubernetes-manager | kubectl apply --dry-run --validate=true -f -
build-all:
	scripts/build-all.sh