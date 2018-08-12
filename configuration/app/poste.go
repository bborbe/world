package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Poste struct {
	Cluster cluster.Cluster
}

func (p *Poste) Childs() []world.Configuration {
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "poste",
		},
	}
}

func (p *Poste) Applier() world.Applier {
	return nil
}

func (p *Poste) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate poste app ...")
	if err := p.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate poste app failed")
	}
	return nil
}
