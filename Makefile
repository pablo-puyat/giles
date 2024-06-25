.DEFAULT_GOAL := build
fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
.PHONY:lint

vet: fmt
	go vet ./...
.PHONY:vet

build: vet
	go build .
.PHONY:build

test:
	find . -name "*.go" | entr -rc go run .
.PHONY:test

build-arm: vet
	CC=x86_64-linux-musl-gcc \
	CXX=x86_64-linux-musl-g++ \
	GOARCH=amd64 \
	GOOS=linux \
	CGO_ENABLED=1 \
	go build -ldflags "-linkmode external -extldflags -static"
