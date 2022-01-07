package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/parser"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Secret struct {
	Key   string
	Value string
}

type GoogleSecretManagerSecret struct {
	Name    string
	Secrets []*Secret
}

type SecretSource struct {
	Key    string `yaml:"key" json:"key"`
	Source string `yaml:"source" json:"source"`
}

type Spec struct {
	Name           string         `yaml:"name" json:"name"`
	Project        string         `yaml:"project" json:"project"`
	SecretsSources []SecretSource `yaml:"secrets" json:"secrets"`
}

type Config struct {
	Spec Spec `yaml:"spec" json:"spec"`
}

func (c *Config) Resolve(client *secretmanager.Client) (*GoogleSecretManagerSecret, error) {
	s := new(GoogleSecretManagerSecret)
	s.Name = c.Spec.Name

	secrets := make([]*Secret, len(c.Spec.SecretsSources))
	for i, source := range c.Spec.SecretsSources {
		secret, err := accessSecretVersion(client, source)
		if err != nil {
			return nil, fmt.Errorf("could not access secret version: %+v", err)
		}
		secrets[i] = secret
	}
	s.Secrets = secrets

	return s, nil
}

func accessSecretVersion(client *secretmanager.Client, source SecretSource) (*Secret, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: source.Source,
	}

	res, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve secret version value: %+v", err)
	} else {
		log.Printf("Response: %+v", res)
	}

	secret := &Secret{
		Key:   source.Key,
		Value: string(res.Payload.Data),
	}
	return secret, nil
}

func main() {
	config := new(Config)

	ctx := context.Background()
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("Could not create Secret Manager client: %+v", err)
	}
	defer c.Close()

	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		templateData, err := config.Resolve(c)
		if err != nil {
			return nil, fmt.Errorf("Could not resolve template data: %+v\n", err)
		}

		template := framework.ResourceTemplate{
			Templates: parser.TemplateStrings(`
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Name }}
  annotations:
    kustomize.config.k8s.io/needs-hash: "true"
  stringData:
    {{range .Secrets -}}
    {{ .Key }}: {{ .Value }}
    {{end}}
			`),
			TemplateData: templateData,
		}

		additions, err := template.Render()
		if err != nil {
			return nil, fmt.Errorf("Could not render secret: %+v\n", err)
		}

		items = append(items, additions...)
		return items, nil
	}

	p := framework.SimpleProcessor{Config: config, Filter: kio.FilterFunc(fn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
