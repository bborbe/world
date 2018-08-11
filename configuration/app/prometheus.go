package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/golang/glog"
)

type Prometheus struct {
	Cluster cluster.Cluster
}

func (p *Prometheus) Childs() []world.Configuration {
	return []world.Configuration{}
}

func (p *Prometheus) Applier() world.Applier {
	return nil
}

func (p *Prometheus) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate prometheus app ...")
	if err := p.Cluster.Validate(ctx); err != nil {
		return err
	}
	return nil
}
