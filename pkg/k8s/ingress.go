package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type IngressApplier struct {
	Context Context
	Ingress Ingress
}

func (s *IngressApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *IngressApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.Ingress,
	}
	return deployer.Apply(ctx)
}

func (s *IngressApplier) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.Ingress.Validate(ctx)
}

type Ingress struct {
	ApiVersion ApiVersion  `yaml:"apiVersion"`
	Kind       Kind        `yaml:"kind"`
	Metadata   Metadata    `yaml:"metadata"`
	Spec       IngressSpec `yaml:"spec"`
}

func (s Ingress) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
}

func (s *Ingress) Validate(ctx context.Context) error {
	if s.ApiVersion != "extensions/v1beta1" {
		return errors.New("invalid ApiVersion")
	}
	if s.Kind != "Ingress" {
		return errors.New("invalid Kind")
	}
	return nil
}

type IngressSpec struct {
	Rules []IngressRule `yaml:"rules"`
}

type IngressHosts []IngressHost

func (i IngressHosts) Validate(ctx context.Context) error {
	if len(i) == 0 {
		return errors.New("IngressHosts empty")
	}
	for _, domain := range i {
		if err := domain.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type IngressHost string

func (i IngressHost) String() string {
	return string(i)
}

func (i IngressHost) Validate(ctx context.Context) error {
	if i == "" {
		return errors.New("ingressHost empty")
	}
	if strings.ContainsRune(i.String(), '_') {
		return errors.New("invalid char in ingressHost")
	}
	return nil
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
