.PHONY: build test clean

BINARY_NAME=disco
BUILD_TAGS=fts5

build:
	go build -tags "$(BUILD_TAGS)" -o $(BINARY_NAME) ./cmd/disco

test:
	go test -tags "$(BUILD_TAGS)" ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f test.db

# Install the binary to $GOPATH/bin
install:
	go install -tags "$(BUILD_TAGS)" ./cmd/disco
