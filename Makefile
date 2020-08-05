build:
	docker build --pull . -t paskalmaksim/kubernetes-manager:dev
test:
	go test *.go
	golangci-lint run
	scripts/validate-license.sh
build-all:
	scripts/build-all.sh