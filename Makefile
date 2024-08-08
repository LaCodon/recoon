.PHONY: all test build docker docker-run

all: test build

test:
	go test -v ./...

build:
	go vet ./...
	go build -o ./bin/recoon ./cmd/recoon
	go build -o ./bin/recoonctl ./cmd/recoonctl

docker:
	docker build -t recoon:dev .

docker-run:
	docker run --rm -it \
 		-v "/var/run/docker.sock:/var/run/docker.sock:rw" \
		-v "${PWD}/test/recooncfg.yaml:/etc/recoon/recooncfg.yaml" \
		-v "${PWD}/test/known_hosts:/etc/recoon/known_hosts" \
		-v "${PWD}/.data:/var/lib/recoon" \
		-p 3680:3680 \
		recoon:dev