.PHONY: build clean test lint mod vendor

ROOT_DIR=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
GO_MAIN_SRC?=$(ROOT_DIR)
GO_CMD=CGO_ENABLED=0 GOOS=linux go
GO_CMD_TEST=CGO_ENABLED=0 go
VERSION=v1.0
GO_APP_NAME=notify

build:
	go build -mod vendor -ldflags "-X main.version=$(VERSION)" -o ${GO_APP_NAME}  $(GO_MAIN_SRC)

clean:
	$(GO_CMD) clean
	rm -rf ${GO_APP_NAME}

test:
	$(GO_CMD_TEST) test ./... -mod=vendor -count=1

lint:
	gofmt -d ./
	go vet ./...
	staticcheck ./...
mod:
	$(GO_CMD) get -u
	$(GO_CMD) mod tidy
	make vendor
vendor:
	$(GO_CMD) mod vendor