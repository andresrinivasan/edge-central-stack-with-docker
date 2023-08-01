.PHONY: build test clean docker

GO=CGO_ENABLED=1 go

# VERSION file is not needed for local development, In the CI/CD pipeline, a temporary VERSION file is written
# if you need a specific version, just override below
# TODO: You need to determine the best approach for setting your service's version for non-development builds.
APPVERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)

# This pulls the version of the SDK from the go.mod file. If the SDK is the only required module,
# it must first remove the word 'required' so the offset of $2 is the same if there are multiple required modules
SDKVERSION=$(shell cat ./go.mod | grep 'github.com/edgexfoundry/app-functions-sdk-go v' | sed 's/require//g' | awk '{print $$2}')

# TODO: Change the service name
MICROSERVICE=simple-app-service
GOFLAGS=-ldflags "-X github.com/edgexfoundry/app-functions-sdk-go/v2/internal.SDKVersion=$(SDKVERSION) -X github.com/edgexfoundry/app-functions-sdk-go/v2/internal.ApplicationVersion=$(APPVERSION)"

# TODO: uncomment and remove default once files are in a Github repository or
#       remove totally including usage below
#GIT_SHA=$(shell git rev-parse HEAD)
GIT_SHA=no-sha

build:
	go mod tidy
	$(GO) build $(GOFLAGS) -o $(MICROSERVICE)

# TODO: Change the registry (iotechsys) below if files are in a Github repository or
#       remove the two lines starting with "-t"
docker:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t iotechsys/simple-app-service:$(GIT_SHA) \
		-t iotechsys/simple-app-service:dev \
		.

test:
	go mod tidy
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]

clean:
	rm -f $(MICROSERVICE)
