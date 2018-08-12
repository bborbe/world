package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Monitoring struct {
	Cluster cluster.Cluster
}

func (m *Monitoring) Childs() []world.Configuration {
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "monitoring",
		},
	}
}

func (m *Monitoring) Applier() world.Applier {
	return nil
}

func (m *Monitoring) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate monitoring app ...")
	if err := m.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate monitoring app failed")
	}
	return nil
}
