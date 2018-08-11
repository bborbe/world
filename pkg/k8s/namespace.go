package k8s

import "fmt"

type NamespaceName string

func (n NamespaceName) String() string {
	return string(n)
}

type Namespace struct {
	ApiVersion ApiVersion `yaml:"apiVersion"`
	Kind       Kind       `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
}

func (n Namespace) String() string {
	return fmt.Sprintf("%s/%s", n.Kind, n.Metadata.Name)
}
