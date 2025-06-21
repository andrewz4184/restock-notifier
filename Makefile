.PHONY: run
.PHONY: build

run:
	go run src/main.go

build:
	GOOS=linux GOARCH=amd64 go build -o bootstrap src/main.go
	zip lambda.zip bootstrap