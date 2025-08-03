.PHONY: run
.PHONY: build
.PHONY: build-docker

run:
	go run src/main.go

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bootstrap src/main.go
	zip lambda.zip bootstrap

build-docker:
	docker run --rm -v "$(PWD)":/var/task -w /var/task golang:1.24.4-alpine sh -c 'apk add --no-cache zip && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bootstrap src/main.go && zip lambda.zip bootstrap'