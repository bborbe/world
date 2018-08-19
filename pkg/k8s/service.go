package k8s

import "fmt"

type Service struct {
	ApiVersion ApiVersion  `yaml:"apiVersion"`
	Kind       Kind        `yaml:"kind"`
	Metadata   Metadata    `yaml:"metadata"`
	Spec       ServiceSpec `yaml:"spec"`
}

func (s Service) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
}

type ServiceSelector map[string]string

type ClusterIP string

type ServiceSpec struct {
	Ports     []Port          `yaml:"ports"`
	Selector  ServiceSelector `yaml:"selector"`
	ClusterIP ClusterIP       `yaml:"clusterIP,omitempty"`
}
