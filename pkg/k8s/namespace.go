package k8s

import "fmt"

type NamespaceName string

type Namespace struct {
	ApiVersion ApiVersion `yaml:"apiVersion"`
	Kind       Kind       `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
}

func (s Namespace) String() string {
	return fmt.Sprintf("%s/%s", s.Kind, s.Metadata.Name)
}
