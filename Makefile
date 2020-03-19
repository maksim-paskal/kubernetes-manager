build:
	docker build . -t paskalmaksim/telegram-gateway:dev
test:
	go test *.go
	golangci-lint run
	scripts/validate-license.sh
build-all:
	scripts/build-all.sh