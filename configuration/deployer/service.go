package deployer

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/pkg/errors"
)

type ServiceDeployer struct {
	Context      world.Context
	Requirements []world.Configuration
	Namespace    k8s.NamespaceName
	Name         k8s.Name
	Ports        []world.Port
}

func (s *ServiceDeployer) Applier() world.Applier {
	return &k8s.Deployer{
		Context: s.Context,
		Data:    s,
	}
}

func (s *ServiceDeployer) Childs() []world.Configuration {
	return s.Requirements
}

func (s *ServiceDeployer) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("Context missing")
	}
	if s.Namespace == "" {
		return errors.New("Namespace missing")
	}
	if s.Name == "" {
		return errors.New("Name missing")
	}
	if len(s.Ports) == 0 {
		return errors.New("Ports missing")
	}
	return nil
}

func (s *ServiceDeployer) Data() (interface{}, error) {
	return s.service(), nil
}

func (s *ServiceDeployer) service() k8s.Service {
	service := k8s.Service{
		ApiVersion: "v1",
		Kind:       "Service",
		Metadata: k8s.Metadata{
			Namespace: s.Namespace,
			Name:      s.Name,
			Labels: k8s.Labels{
				"app": s.Name.String(),
			},
		},
		Spec: k8s.ServiceSpec{
			Selector: k8s.ServiceSelector{
				"app": s.Name.String(),
			},
		},
	}
	for _, port := range s.Ports {
		service.Spec.Ports = append(service.Spec.Ports, k8s.Port{
			Name: k8s.PortName(port.Name),
			Port: k8s.PortNumber(port.Port),
		})
	}
	return service
}
