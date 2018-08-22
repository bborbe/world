package deployer

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type ServiceDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.Name
	Requirements []world.Configuration
	Ports        []Port
	ClusterIP    k8s.ClusterIP
	Labels       k8s.Labels
	Annotations  k8s.Annotations
}

func (s *ServiceDeployer) Applier() (world.Applier, error) {
	return &k8s.ServiceApplier{
		Context: s.Context,
		Service: s.service(),
	}, nil
}

func (s *ServiceDeployer) Children() []world.Configuration {
	return s.Requirements
}

func (s *ServiceDeployer) service() k8s.Service {
	labels := k8s.Labels{}
	for k, v := range s.Labels {
		labels[k] = v
	}
	labels["app"] = s.Name.String()
	service := k8s.Service{
		ApiVersion: "v1",
		Kind:       "Service",
		Metadata: k8s.Metadata{
			Namespace:   s.Namespace,
			Name:        s.Name,
			Labels:      labels,
			Annotations: s.Annotations,
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
