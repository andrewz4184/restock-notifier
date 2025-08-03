#!/bin/bash
# Build Go Lambda using Docker to match AWS Lambda environment

docker run --rm \
  -v "$PWD":/var/task \
  -w /var/task \
  golang:1.22-alpine \
  sh -c '
    apk add --no-cache zip
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap src/main.go
    zip lambda.zip bootstrap
  '
