.PHONY: test build

test:
	go test ./...

build:
	go build -o bin/engine ./cmd/cli
	go build -o bin/engine-ui ./cmd/ui
	go build -o bin/engine-encrypt ./cmd/encrypt
