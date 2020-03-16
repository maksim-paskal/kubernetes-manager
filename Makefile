build:
	docker build .
test:
	golangci-lint run