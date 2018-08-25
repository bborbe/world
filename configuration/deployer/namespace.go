package deployer

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type NamespaceDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Requirements []world.Configuration
}

func (w *NamespaceDeployer) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.Context,
		w.Namespace,
	)
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
			Name:      k8s.MetadataName(n.Namespace),
			Namespace: n.Namespace,
			Labels: k8s.Labels{
				"app": n.Namespace.String(),
			},
		},
	}
}
