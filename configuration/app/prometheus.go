package app

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
)

type Prometheus struct {
	Cluster cluster.Cluster
}

func (p *Prometheus) Children() []world.Configuration {
	return []world.Configuration{}
}

func (p *Prometheus) Applier() (world.Applier, error) {
	return nil, nil
}
