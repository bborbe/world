package deployer

import (
	"context"

	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type IngressDeployer struct {
	Context      world.Context
	Requirements []world.Configuration
	Namespace    world.Namespace
	Domains      []world.Domain
}

func (i *IngressDeployer) Applier() world.Applier {
	return &k8s.Deployer{
		Context: i.Context,
		Data:    i.ingress(),
	}
}

func (i *IngressDeployer) Childs() []world.Configuration {
	return i.Requirements
}

func (i *IngressDeployer) Validate(ctx context.Context) error {
	if i.Context == "" {
		return fmt.Errorf("Context missing")
	}
	if i.Namespace == "" {
		return fmt.Errorf("Namespace missing")
	}
	if len(i.Domains) == 0 {
		return fmt.Errorf("Domains missing")
	}
	return nil
}

func (i *IngressDeployer) ingress() k8s.Ingress {
	ingress := k8s.Ingress{
		ApiVersion: "extensions/v1beta1",
		Kind:       "Ingress",
		Metadata: k8s.Metadata{
			Namespace: k8s.NamespaceName(i.Namespace),
			Name:      k8s.Name(i.Namespace),
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
			Host: k8s.IngressHost(domain),
			Http: k8s.IngressHttp{
				Paths: []k8s.IngressPath{
					{
						Path: "/",
						Backends: k8s.IngressBackend{
							ServiceName: k8s.IngressBackendServiceName(i.Namespace),
							ServicePort: "web",
						},
					},
				},
			},
		})
	}
	return ingress
}
