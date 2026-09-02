.PHONY: proto fmt test build

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative desc/travel.proto

fmt:
	gofmt -w $$(find . -name '*.go')

test:
	go test ./...

build:
	go build ./...
