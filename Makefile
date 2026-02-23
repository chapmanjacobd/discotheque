.PHONY: build test cover webtest webcover clean fmt lint sql install all readme dev

BINARY_NAME=disco
BUILD_TAGS=fts5

all: fmt lint sql test webtest build readme

build:
	go build -tags "$(BUILD_TAGS)" -o $(BINARY_NAME) ./cmd/disco

dev:
	(sleep 2 && xdg-open http://localhost:5555) &
	air -d

readme: build
	./$(BINARY_NAME) readme > README.md

test:
	go test -tags "$(BUILD_TAGS)" -coverprofile=coverage.out ./...

cover: test
	go tool cover -func=coverage.out | awk '{n=split($$NF,a,"%%"); if (a[1] < 85) print $$0}' | sort -k3 -n

webtest:
	npm test --prefix web

webcover:
	npm run cover --prefix web

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
