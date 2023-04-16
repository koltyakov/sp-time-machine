funcRoot := ./functions
srvPath  := ./functions/bin/server

ifneq (,$(wildcard ./.env))
	include .env
	export
endif

install:
	go get -u ./... && go mod tidy

format:
	gofmt -s -w .

build:
	go build -o bin/wfsync ./cmd/worker/...
	cp config.jsonc bin/config.jsonc

test:
	go test ./... -v -race -count=1

run:
	go run .

build-fns: clean
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -tags "prod" -o $(funcRoot)/bin/server ./cmd/server/...
	cp config.jsonc $(funcRoot)/config.jsonc

pack-fns: build-fns
	@mkdir -p infra/package
	cd $(funcRoot) && func pack -o ../infra/package/functions

start-fns:
	@go build -tags "prod" -o $(funcRoot)/bin/server ./cmd/server/...
	@cd $(funcRoot) && func start # --verbose

publish: build-fns
	cd $(funcRoot) && func azure functionapp publish ${AZURE_FUNCTIONS_APP}

clean:
	rm -rf bin/ tmp/ $(funcRoot)/bin/ $(funcRoot)/tmp/ infra/package/

terra:
	cd infra && make terra
