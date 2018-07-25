package k8s

type Service struct {
	ApiVersion ApiVersion  `yaml:"apiVersion"`
	Kind       Kind        `yaml:"kind"`
	Metadata   Metadata    `yaml:"metadata"`
	Spec       ServiceSpec `yaml:"spec"`
}

type ServiceSelector map[string]string

type ServiceSpec struct {
	Ports    []Port          `yaml:"ports"`
	Selector ServiceSelector `yaml:"selector"`
}
