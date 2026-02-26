.PHONY: build test cover webtest webcover clean fmt lint sql install all readme dev

BINARY_NAME=disco
SYNCWEB_BINARY=syncweb

MIN_TAGS=noassets
BUILD_TAGS=noassets,fts5
SYNCWEB_TAGS=$(BUILD_TAGS),syncweb

all: fmt lint sql test webtest build readme

build:
	go build -tags "$(BUILD_TAGS)" -o $(BINARY_NAME) ./cmd/disco
	go build -tags "$(SYNCWEB_TAGS)" -o $(SYNCWEB_BINARY) ./cmd/syncweb

dev:
	(sleep 2 && xdg-open http://localhost:5555) &
	air -d

readme: build
	./$(BINARY_NAME) readme > README.md

test:
	go test -tags "$(MIN_TAGS)" ./...
	go test -tags "$(BUILD_TAGS)" ./...
	go test -tags "$(SYNCWEB_TAGS)" ./...

cover:
	go test -tags "$(SYNCWEB_TAGS)" -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | awk '{n=split($$NF,a,"%%"); if (a[1] < 85) print $$0}' | sort -k3 -n

webtest:
	npm test --prefix web

webcover:
	npm run cover --prefix web

fmt:
	gofmt -s -w -e .
	go fix -tags "$(BUILD_TAGS)" ./...
	-goimports -w -e .
	-gofumpt -w .
	-gci write .

lint:
	-staticcheck -tags "$(BUILD_TAGS)" ./...
	go vet -tags "$(BUILD_TAGS)" ./...

sql:
	sqlc generate
	sqlc vet
	-sqlc diff

clean:
	rm -f $(BINARY_NAME) $(SYNCWEB_BINARY)
	rm -f test.db
	rm -f coverage.out

# Install the binary to $GOPATH/bin
install:
	go install -tags "$(BUILD_TAGS)" ./cmd/disco
	go install -tags "$(BUILD_TAGS)" ./cmd/syncweb
