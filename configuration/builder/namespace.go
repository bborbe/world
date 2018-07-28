package builder

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type NamespaceBuilder struct {
	Namespace world.Namespace
}

func (n *NamespaceBuilder) Build() k8s.Namespace {
	return k8s.Namespace{
		ApiVersion: "v1",
		Kind:       "Namespace",
		Metadata: k8s.Metadata{
			Name: k8s.Name(n.Namespace),
			Labels: k8s.Labels{
				"app": n.Namespace.String(),
			},
		},
	}
}
