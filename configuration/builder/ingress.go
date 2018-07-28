package builder

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type IngressBuilder struct {
	Namespace world.Namespace
	Domains   []world.Domain
}

func (i *IngressBuilder) Build() k8s.Ingress {
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
