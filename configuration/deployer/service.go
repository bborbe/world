package deployer

import (
	"context"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
	"github.com/pkg/errors"
)

type ServiceDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.MetadataName
	Ports        []Port
	ClusterIP    k8s.ClusterIP
	Labels       k8s.Labels
	Annotations  k8s.Annotations
	Requirements []world.Configuration
}

func (t *ServiceDeployer) Validate(ctx context.Context) error {
	if len(t.Ports) == 0 {
		return errors.New("service has no ports")
	}
	for _, port := range t.Ports {
		if err := port.Validate(ctx); err != nil {
			return errors.Wrap(err, "Port invalid")
		}
	}
	return validation.Validate(
		ctx,
		t.Context,
		t.Namespace,
		t.Name,
	)
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
			ClusterIP: s.ClusterIP,
		},
	}
	for _, port := range s.Ports {
		service.Spec.Ports = append(service.Spec.Ports, k8s.Port{
			Name:     port.Name,
			Port:     port.Port,
			Protocol: port.Protocol,
		})
	}
	return service
}
