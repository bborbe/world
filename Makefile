
all: test install

install:
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install cmd/world-apply/*.go

test:
	go test -cover -race $(shell go list ./... | grep -v /vendor/)

format:
	go get golang.org/x/tools/cmd/goimports
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	find . -type f -name '*.go' -not -path './vendor/*' -exec goimports -w "{}" +

generate:
	go get github.com/maxbrunsfeld/counterfeiter
	go generate ./...
