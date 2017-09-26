BINARY=ms-sendgrid-webhook
VERSION:=0.1.0

.PHONY: all
.DEFAULT_GOAL := all

test:
	go test  -v ./...

go-stub:
	 protoc -I followservice/ followservice/follow.proto --go_out=plugins=grpc:followservice

get:
	go get

docker:
	docker build -t clicrdv/${BINARY}:${VERSION} .

tag-latest:
	docker tag clicrdv/${BINARY}:${VERSION} clicrdv/${BINARY}:latest

binary:
	go build -o ${BINARY}-osx main.go
	env GOOS=linux GOARCH=amd64 go build -o ${BINARY}-linux main.go

all: get test go-stub binary docker tag-latest
