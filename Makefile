BINARY := cred

.PHONY: all build install clean test test-integration lint fmt tidy help

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
	@echo "  make fmt     - Run gofmt"
	@echo "  make tidy    - Run go mod tidy"
	@echo "  make clean   - Remove binary"
