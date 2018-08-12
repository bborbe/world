package deployer

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/pkg/errors"
)

type IngressDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.Name
	Requirements []world.Configuration
	Domains      []world.Domain
}

func (i *IngressDeployer) Applier() world.Applier {
	return &k8s.Deployer{
		Context: i.Context,
		Data:    i,
	}
}

func (i *IngressDeployer) Childs() []world.Configuration {
	return i.Requirements
}

func (i *IngressDeployer) Validate(ctx context.Context) error {
	if i.Context == "" {
		return errors.New("Context missing in ingress deployer")
	}
	if i.Namespace == "" {
		return errors.New("Namespace missing in ingress deployer")
	}
	if i.Name == "" {
		return errors.New("Name missing in ingress deployer")
	}
	if len(i.Domains) == 0 {
		return errors.New("Domains missing in ingress deployer")
	}
	return nil
}

func (i *IngressDeployer) Data() (interface{}, error) {
	return i.ingress(), nil
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
			Host: k8s.IngressHost(domain),
			Http: k8s.IngressHttp{
				Paths: []k8s.IngressPath{
					{
						Path: "/",
						Backends: k8s.IngressBackend{
							ServiceName: k8s.IngressBackendServiceName(i.Name),
							ServicePort: "web",
						},
					},
				},
			},
		})
	}
	return ingress
}
