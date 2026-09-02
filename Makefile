# Keep the tourism RPC generation flow aligned with merchant-rpc/simple-admin tooling.
SERVICE=Travel
SERVICE_STYLE=travel
SERVICE_LOWER=travel
SERVICE_DASH=travel
VERSION=0.1.0
PROJECT_STYLE=go_zero
PROJECT_BUILD_SUFFIX=rpc
ENT_FEATURE=sql/execquery
GOARCH=amd64

.PHONY: test fmt gen-rpc gen-ent build-linux

test:
	go test -v ./...

fmt:
	gofmt -w $$(find . -name '*.go')

gen-rpc:
	goctls rpc protoc ./desc/travel.proto --go_out=./travel --go-grpc_out=./travel --zrpc_out=. --style=$(PROJECT_STYLE)

# Generate Ent from the authoritative schemas. Do not hand-write generated files.
gen-ent:
	go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema --feature $(ENT_FEATURE)

build-linux:
	env CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -trimpath -o travel_rpc travel.go
