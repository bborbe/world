package k8s

import "fmt"

type Secret struct {
	ApiVersion ApiVersion `yaml:"apiVersion"`
	Kind       Kind       `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
	Type       SecretType `yaml:"type"`
	Data       SecretData `yaml:"data"`
}

func (s Secret) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
}

type SecretType string

type SecretData map[string]string
