package deployer

import (
	"context"

	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type ServiceDeployer struct {
	Context      world.Context
	Requirements []world.Configuration
	Namespace    world.Namespace
	Ports        []world.Port
}

func (s *ServiceDeployer) Applier() world.Applier {
	return &k8s.Deployer{
		Context: s.Context,
		Data:    s.service(),
	}
}

func (s *ServiceDeployer) Childs() []world.Configuration {
	return s.Requirements
}

func (s *ServiceDeployer) Validate(ctx context.Context) error {
	if s.Context == "" {
		return fmt.Errorf("Context missing")
	}
	if s.Namespace == "" {
		return fmt.Errorf("Namespace missing")
	}
	if len(s.Ports) == 0 {
		return fmt.Errorf("Ports missing")
	}
	return nil
}

func (s *ServiceDeployer) service() k8s.Service {
	service := k8s.Service{
		ApiVersion: "v1",
		Kind:       "Service",
		Metadata: k8s.Metadata{
			Namespace: k8s.NamespaceName(s.Namespace),
			Name:      k8s.Name(s.Namespace),
			Labels: k8s.Labels{
				"app": s.Namespace.String(),
			},
		},
		Spec: k8s.ServiceSpec{

			Selector: k8s.ServiceSelector{
				"app": s.Namespace.String(),
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
