BINARY := cred

GOVULNCHECK_VERSION := v1.9.0
GOVULNCHECK         := go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

.PHONY: all build install clean test test-integration lint vulncheck fmt tidy help

all: build

build: tidy
	go build -o $(BINARY) .

install: tidy
	go install .

clean:
	rm -f $(BINARY)

test:
	go test ./...

test-integration:
	go test -tags integration -v -count=1 ./integration/

lint:
	go vet ./...

# Not part of lint: the result depends on that day's vulnerability database,
# not on the code. CI runs it via .github/workflows/govulncheck.yml.
# The standard library is checked against the Go version running this.
vulncheck:
	$(GOVULNCHECK) ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

help:
	@echo "Usage:"
	@echo "  make build   - Build binary (go mod tidy + go build)"
	@echo "  make install - Install to GOPATH/bin (go install)"
	@echo "  make test    - Run tests"
	@echo "  make lint    - Run go vet"
	@echo "  make vulncheck - Check for known vulnerabilities (govulncheck)"
	@echo "  make fmt     - Run gofmt"
	@echo "  make tidy    - Run go mod tidy"
	@echo "  make clean   - Remove binary"
