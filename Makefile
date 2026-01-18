.PHONY: build test clean install

BINARY_NAME=atom
BUILD_DIR=./cmd/atom

build:
	go build -o $(BINARY_NAME) $(BUILD_DIR)

test:
	go test ./... -v

clean:
	rm -f $(BINARY_NAME)
	go clean

install: build
	go install $(BUILD_DIR)

lint:
	go vet ./...
