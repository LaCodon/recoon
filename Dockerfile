FROM golang:1.19-alpine3.17 AS builder

WORKDIR /code

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/recoon ./cmd/recoon

FROM alpine:3.17

# volume for data (repos, bbolt)
VOLUME /var/lib/recoon

ENV SSH_KNOWN_HOSTS="/etc/recoon/known_hosts"

# UI connection port
EXPOSE 3680

RUN apk update && apk add docker docker-cli-compose

COPY --from=builder /bin/recoon /recoon

ENTRYPOINT /recoon
