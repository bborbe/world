package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/golang/glog"
)

type Proxy struct {
	Cluster cluster.Cluster
}

func (p *Proxy) Childs() []world.Configuration {
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
		return err
	}
	return nil
}
