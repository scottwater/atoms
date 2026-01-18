.PHONY: build test clean install

BINARY_NAME=atom
BUILD_DIR=./cmd/atom
OUT_DIR=bin

build:
	@mkdir -p $(OUT_DIR)
	go build -o $(OUT_DIR)/$(BINARY_NAME) $(BUILD_DIR)

test:
	go test ./... -v

clean:
	rm -rf $(OUT_DIR)
	go clean

install: build
	go install $(BUILD_DIR)

lint:
	go vet ./...
