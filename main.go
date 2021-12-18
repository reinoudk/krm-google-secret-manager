package main

import (
	"fmt"
	"os"
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
	Secrets []Secret
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

func (c *Config) Resolve() (*GoogleSecretManagerSecret, error) {
	s := new(GoogleSecretManagerSecret)
	s.Name = c.Spec.Name

	secrets := make([]Secret, len(c.Spec.SecretsSources))
	for i, source := range c.Spec.SecretsSources {
		secret := Secret{
			Key:   source.Key,
			Value: source.Source,
		}
		secrets[i] = secret
	}
	s.Secrets = secrets

	return s, nil
}

func main() {
	config := new(Config)

	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		templateData, err := config.Resolve()
		if err != nil {
			return nil, fmt.Errorf("Could not resolve template data: %+v\n", err)
		}

		template := framework.ResourceTemplate{
			Templates: parser.TemplateStrings(`
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Name }}
  stringData:
  {{ range .Secrets }}- {{ .Key }}: {{ .Value }}{{end}}
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
