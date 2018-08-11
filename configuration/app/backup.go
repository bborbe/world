package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/golang/glog"
)

type Backup struct {
	Cluster cluster.Cluster
}

func (b *Backup) Childs() []world.Configuration {
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   b.Cluster.Context,
			Namespace: "backup",
		},
	}
}

func (b *Backup) Applier() world.Applier {
	return nil
}

func (b *Backup) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate backup app ...")
	if err := b.Cluster.Validate(ctx); err != nil {
		return err
	}
	return nil
}
