TAG ?= v0.1.0-local

default: build-docker

# Build docker image
build-docker:
	docker build -t ghcr.io/reinoudk/krm-google-secret-manager:$(TAG) .

# Run example
example:
	kubectl kustomize --enable-alpha-plugins example

# Generate Dockerfile
gen:
	go run ./main.go gen ./

.PHONY: example
