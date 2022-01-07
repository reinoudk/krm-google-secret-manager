# krm-google-secret-manager

This is a [KRM Function](https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md)
that generates Kubernetes Secrets from Google Secret Manager secret versions.

## Usage

Due to [mounting issues](https://github.com/kubernetes-sigs/kustomize/issues/4290) with containerized
functions, the generator should be run as an exec function (for now). The Google credentials for fetching the
secret version are found using [Application Default Credentials](https://cloud.google.com/docs/authentication/production#automatically). 

```yaml
apiVersion: kustomize.reinoud.dev/v1
kind: GoogleSecretManagerSecretGenerator
metadata:
  name: not-important
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ../../build/krm-google-secret-manager
spec:
  name: example
  project:
  secrets:
    - key: example-key
      source: projects/<your-project>/secrets/<your-secret>/versions/latest
```

See more details in [examples/exec](example/exec)

## Building the function

Simply call `make` to build the function and store the binary in `build/`.
