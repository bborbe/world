package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Proxy struct {
	Cluster cluster.Cluster
}

func (p *Proxy) Children() []world.Configuration {
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   p.Cluster.Context,
			Namespace: "proxy",
		},
	}
}

func (p *Proxy) Applier() world.Applier {
	return nil
}

func (p *Proxy) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate proxy app ...")
	if err := p.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate proxy app failed")
	}
	return nil
}
