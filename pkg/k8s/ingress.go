package k8s

import "fmt"

type Ingress struct {
	ApiVersion ApiVersion  `yaml:"apiVersion"`
	Kind       Kind        `yaml:"kind"`
	Metadata   Metadata    `yaml:"metadata"`
	Spec       IngressSpec `yaml:"spec"`
}

func (s Ingress) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
}

type IngressSpec struct {
	Rules []IngressRule `yaml:"rules"`
}

type IngressHost string

func (i IngressHost) String() string {
	return string(i)
}

type IngressRule struct {
	Host IngressHost `yaml:"host"`
	Http IngressHttp `yaml:"http"`
}

type IngressHttp struct {
	Paths []IngressPath `yaml:"paths"`
}

type IngressPathPath string

type IngressPath struct {
	Backends IngressBackend  `yaml:"backend"`
	Path     IngressPathPath `yaml:"path"`
}

type IngressBackendServiceName string

type IngressBackendServicePort string

type IngressBackend struct {
	ServiceName IngressBackendServiceName `yaml:"serviceName"`
	ServicePort IngressBackendServicePort `yaml:"servicePort"`
}
