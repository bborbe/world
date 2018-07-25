package k8s

type Name string

type Metadata struct {
	Namespace   NamespaceName `yaml:"namespace,omitempty"`
	Name        Name          `yaml:"name,omitempty"`
	Labels      Labels        `yaml:"labels,omitempty"`
	Annotations Annotations   `yaml:"annotations,omitempty"`
}
