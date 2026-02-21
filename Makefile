.PHONY: build test clean fmt lint sql install all readme dev

BINARY_NAME=disco
BUILD_TAGS=fts5

all: fmt lint sql test build readme

build:
	go build -tags "$(BUILD_TAGS)" -o $(BINARY_NAME) ./cmd/disco

dev:
	(sleep 2 && xdg-open http://localhost:5555) &
	air -d

readme: build
	./$(BINARY_NAME) readme > README.md

test:
	go test -tags "$(BUILD_TAGS)" -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep -v "100.0%" | sort -k3 -n

fmt:
	gofmt -s -w -e .
	go fix ./...
	-goimports -w -e .
	-gofumpt -w .
	-gci write .

lint:
	-staticcheck ./...
	go vet ./...

sql:
	sqlc generate
	sqlc vet
	-sqlc diff

clean:
	rm -f $(BINARY_NAME)
	rm -f test.db
	rm -f coverage.out

# Install the binary to $GOPATH/bin
install:
	go install -tags "$(BUILD_TAGS)" ./cmd/disco
