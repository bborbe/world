
all: test install

install:
	GOBIN=$(GOPATH)/bin go install .

deps:
	go get -u github.com/kisielk/errcheck
	go get -u github.com/maxbrunsfeld/counterfeiter/v6
	go get -u github.com/onsi/ginkgo/ginkgo
	go get -u golang.org/x/lint/golint
	go get -u golang.org/x/tools/cmd/goimports

precommit: ensure format generate test check addlicense
	@echo "ready to commit"

format:
	@GO111MODULE=off go get github.com/seibert-media/goimports-reviser
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	find . -type f -name '*.go' -not -path './vendor/*' -exec goimports-reviser -project-name github.com/bborbe/world -file-path "{}" \;

ensure:
	GO111MODULE=on go mod verify
	GO111MODULE=on go mod vendor

generate:
	GO111MODULE=on go get github.com/maxbrunsfeld/counterfeiter/v6
	rm -rf mocks
	GO111MODULE=on go generate ./...

test:
	GO111MODULE=on go test -cover -race $(shell go list ./... | grep -v /vendor/)

check: lint vet errcheck

lint:
	@GO111MODULE=on go get golang.org/x/lint/golint
	@golint -min_confidence 1 $(shell go list ./... | grep -v /vendor/)

vet:
	@GO111MODULE=on go vet $(shell go list ./... | grep -v /vendor/)

errcheck:
	@GO111MODULE=on go get github.com/kisielk/errcheck
	@errcheck -ignore '(Close|Write|Fprint)' $(shell go list ./... | grep -v /vendor/)

addlicense:
	@GO111MODULE=on go get github.com/google/addlicense
	@addlicense -c "Benjamin Borbe" -y 2020 -l bsd ./*.go ./pkg/*/*.go ./configuration/*.go  ./configuration/*/*.go
