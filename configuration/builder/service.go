package builder

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type ServiceBuilder struct {
	Namespace world.Namespace
	Port      world.Port
}

func (s *ServiceBuilder) Build() k8s.Service {
	return k8s.Service{
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
			Ports: []k8s.Port{
				{
					Name:       "web",
					Port:       k8s.PortNumber(s.Port),
					Protocol:   "TCP",
					TargetPort: "http",
				},
			},
			Selector: k8s.ServiceSelector{
				"app": s.Namespace.String(),
			},
		},
	}
}
