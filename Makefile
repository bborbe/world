
all: precommit install

install:
	GOBIN=$(GOPATH)/bin go install .

precommit: ensure format generate test check addlicense
	@echo "ready to commit"

format:
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	find . -type f -name '*.go' -not -path './vendor/*' -exec go run -mod=vendor github.com/incu6us/goimports-reviser -project-name github.com/bborbe/world -file-path "{}" \;

ensure:
	go mod verify
	go mod vendor

generate:
	rm -rf mocks
	go generate ./...

test:
	go test -mod=vendor -p=1 -cover -race $(shell go list -mod=vendor ./... | grep -v /vendor/)

check: lint vet errcheck vulncheck

lint:
	go run -mod=vendor golang.org/x/lint/golint -min_confidence 1 $(shell go list -mod=vendor ./... | grep -v /vendor/)

vet:
	go vet -mod=vendor $(shell go list -mod=vendor ./... | grep -v /vendor/)

errcheck:
	go run -mod=vendor github.com/kisielk/errcheck -ignore '(Close|Write|Fprint)' $(shell go list -mod=vendor ./... | grep -v /vendor/)

vulncheck:
	go run -mod=vendor golang.org/x/vuln/cmd/govulncheck $(shell go list -mod=vendor ./... | grep -v /vendor/)

addlicense:
	go run -mod=vendor github.com/google/addlicense -c "Benjamin Borbe" -y $$(date +'%Y') -l bsd $$(find . -name "*.go" -not -path './vendor/*')
