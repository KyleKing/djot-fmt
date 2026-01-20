.PHONY: help build test install clean fmt lint

help:
	@echo "Available targets:"
	@echo "  build    - Build the binary"
	@echo "  test     - Run tests"
	@echo "  install  - Install binary to GOPATH/bin"
	@echo "  clean    - Remove build artifacts"
	@echo "  fmt      - Format code"
	@echo "  lint     - Run linters"

build:
	go build -o djot-fmt

test:
	go test -v ./...

install:
	go install

clean:
	rm -f djot-fmt
	go clean

fmt:
	hk fix

lint:
	hk check
