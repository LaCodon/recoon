.PHONY: test build all

all: test build

test:
	go test -v ./...

build:
	go vet ./...
	go build -o ./bin/recoon ./cmd/recoon
	go build -o ./bin/recoonctl ./cmd/recoonctl