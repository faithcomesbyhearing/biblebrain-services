.PHONY: build clean deploy undeploy offline gomodgen create-build precommit standardize

precommit: clean standardize build
	golangci-lint run

standardize: 
	goimports -w .
	gofumpt -w .

build: gomodgen create-build

clean:
	rm -rf ./bin ./vendor 

offline: build
	yarn sls offline --stage ${environment} --httpPort 3008 --noTimeout

deploy: clean build
	sls deploy --stage ${environment} 

undeploy:
	sls remove

gomodgen:
	export GO111MODULE=on
	go mod tidy

GO_ENV=GOARCH=${GOARCH} GOOS=${GOOS} CGO_ENABLED=0

create-build:
	env ${GO_ENV} go build ${GOFLAGS} -o bin/status status/main.go
	env ${GO_ENV} go build ${GOFLAGS} -o bin/bservice cmd/httpserver/api/main.go

vet:
	go vet ./...
