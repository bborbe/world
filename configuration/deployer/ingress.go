package deployer

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type IngressDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.Name
	Requirements []world.Configuration
	Domains      []k8s.IngressHost
	Port         k8s.PortName
}

func (i *IngressDeployer) Applier() (world.Applier, error) {
	return &k8s.IngressApplier{
		Context: i.Context,
		Ingress: i.ingress(),
	}, nil
}

func (i *IngressDeployer) Children() []world.Configuration {
	return i.Requirements
}

func (i *IngressDeployer) ingress() k8s.Ingress {
	ingress := k8s.Ingress{
		ApiVersion: "extensions/v1beta1",
		Kind:       "Ingress",
		Metadata: k8s.Metadata{
			Namespace: i.Namespace,
			Name:      i.Name,
			Labels: k8s.Labels{
				"app": i.Namespace.String(),
			},
			Annotations: k8s.Annotations{
				"kubernetes.io/ingress.class": "traefik",
				"traefik.frontend.priority":   "10000",
			},
		},
		Spec: k8s.IngressSpec{},
	}
	for _, domain := range i.Domains {
		ingress.Spec.Rules = append(ingress.Spec.Rules, k8s.IngressRule{
			Host: domain,
			Http: k8s.IngressHttp{
				Paths: []k8s.IngressPath{
					{
						Path: "/",
						Backends: k8s.IngressBackend{
							ServiceName: k8s.IngressBackendServiceName(i.Name),
							ServicePort: k8s.IngressBackendServicePort(i.Port),
						},
					},
				},
			},
		})
	}
	return ingress
}
