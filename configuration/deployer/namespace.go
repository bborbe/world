package deployer

import (
	"context"

	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type NamespaceDeployer struct {
	Context      world.Context
	Requirements []world.Configuration
	Namespace    world.Namespace
}

func (n *NamespaceDeployer) Applier() world.Applier {
	return &k8s.Deployer{
		Context: n.Context,
		Data:    n.namespace(),
	}
}

func (n *NamespaceDeployer) Childs() []world.Configuration {
	return n.Requirements
}

func (n *NamespaceDeployer) Validate(ctx context.Context) error {
	if n.Context == "" {
		return fmt.Errorf("Context missing")
	}
	if n.Namespace == "" {
		return fmt.Errorf("Namespace missing")
	}
	return nil
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
