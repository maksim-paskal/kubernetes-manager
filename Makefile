build:
	docker build .
test:
	go test *.go
	golangci-lint run
	scripts/validate-license.sh