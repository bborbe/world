package deployer

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/pkg/errors"
)

type NamespaceDeployer struct {
	Context      k8s.Context
	Requirements []world.Configuration
	Namespace    k8s.NamespaceName
}

func (n *NamespaceDeployer) Applier() world.Applier {
	return &k8s.Deployer{
		Context: n.Context,
		Data:    n,
	}
}

func (n *NamespaceDeployer) Children() []world.Configuration {
	return n.Requirements
}

func (n *NamespaceDeployer) Validate(ctx context.Context) error {
	if n.Context == "" {
		return errors.New("Context missing in namespace deployer")
	}
	if n.Namespace == "" {
		return errors.New("Namespace missing in namespace deployer")
	}
	return nil
}

func (n *NamespaceDeployer) Data() (interface{}, error) {
	return n.namespace(), nil
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
