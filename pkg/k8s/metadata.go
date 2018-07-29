package k8s

import "fmt"

type Name string

type Metadata struct {
	Namespace   NamespaceName `yaml:"namespace,omitempty"`
	Name        Name          `yaml:"name,omitempty"`
	Labels      Labels        `yaml:"labels,omitempty"`
	Annotations Annotations   `yaml:"annotations,omitempty"`
}

func (m Metadata) String() string {
	return fmt.Sprintf("ns: %s name: %s", m.Namespace, m.Name)
}
