default: build

# Build a local binary
build:
	CGO_ENABLED=0 go build -o build/krm-google-secret-manager

# Run example
example: build
	kustomize build --enable-alpha-plugins --enable-exec examples/exec

# Generate Dockerfile
gen:
	go run ./main.go gen ./

.PHONY: example build
