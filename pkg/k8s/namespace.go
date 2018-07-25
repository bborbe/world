package k8s

type NamespaceName string

type Namespace struct {
	ApiVersion ApiVersion `yaml:"apiVersion"`
	Kind       Kind       `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
}
