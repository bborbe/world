package k8s

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

type ServiceApplier struct {
	Context Context
	Service Service
}

func (s *ServiceApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *ServiceApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.Service,
	}
	return deployer.Apply(ctx)
}

func (s *ServiceApplier) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.Service.Validate(ctx)
}

type Service struct {
	ApiVersion ApiVersion  `yaml:"apiVersion"`
	Kind       Kind        `yaml:"kind"`
	Metadata   Metadata    `yaml:"metadata"`
	Spec       ServiceSpec `yaml:"spec"`
}

func (s Service) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
}

func (c *Service) Validate(ctx context.Context) error {
	return nil
}

type ServiceSelector map[string]string

type ClusterIP string

func (s ClusterIP) String() string {
	return string(s)
}

func (s ClusterIP) Validate(ctx context.Context) error {
	if s == "" {
		return errors.New("ClusterIP missing")
	}
	return nil
}

type ClusterName string

func (s ClusterName) Validate(ctx context.Context) error {
	if s == "" {
		return errors.New("ClusterName missing")
	}
	return nil
}

type ServiceSpec struct {
	Ports     []Port          `yaml:"ports"`
	Selector  ServiceSelector `yaml:"selector"`
	ClusterIP ClusterIP       `yaml:"clusterIP,omitempty"`
}
