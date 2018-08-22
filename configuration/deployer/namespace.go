package deployer

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type NamespaceDeployer struct {
	Context      k8s.Context
	Requirements []world.Configuration
	Namespace    k8s.NamespaceName
}

func (n *NamespaceDeployer) Applier() (world.Applier, error) {
	return &k8s.NamespaceApplier{
		Context:   n.Context,
		Namespace: n.namespace(),
	}, nil
}

func (n *NamespaceDeployer) Children() []world.Configuration {
	return n.Requirements
}

func (n *NamespaceDeployer) namespace() k8s.Namespace {
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
