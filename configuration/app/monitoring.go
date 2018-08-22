package app

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
)

type Monitoring struct {
	Cluster cluster.Cluster
}

func (m *Monitoring) Children() []world.Configuration {
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "monitoring",
		},
	}
}

func (m *Monitoring) Applier() (world.Applier, error) {
	return nil, nil
}
