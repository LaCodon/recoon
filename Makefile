.PHONY: build

build:
	go build -o ./bin/recoon ./cmd/recoon
	go build -o ./bin/recoonctl ./cmd/recoonctl