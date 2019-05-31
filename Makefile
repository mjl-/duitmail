export CGO_ENABLED=0
export GOPROXY=off
export GOFLAGS=-mod=vendor

build:
	go build
	go vet
	golint

install:
	go install

clean:
	go clean

fmt:
	go fmt ./...
