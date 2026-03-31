BINARY := cred
MODULE := github.com/mitsuru/cred

.PHONY: all build clean test lint fmt tidy help

all: build

build: tidy
	go build -o $(BINARY) .

clean:
	rm -f $(BINARY)

test:
	go test ./...

lint:
	go vet ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

help:
	@echo "Usage:"
	@echo "  make build  - ビルド (go mod tidy + go build)"
	@echo "  make test   - テスト実行"
	@echo "  make lint   - go vet 実行"
	@echo "  make fmt    - gofmt でフォーマット"
	@echo "  make tidy   - go mod tidy 実行"
	@echo "  make clean  - バイナリ削除"
