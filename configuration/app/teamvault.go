package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/golang/glog"
)

type Teamvault struct {
	Cluster cluster.Cluster
}

func (t *Teamvault) Childs() []world.Configuration {
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   t.Cluster.Context,
			Namespace: "teamvault",
		},
	}
}

func (t *Teamvault) Applier() world.Applier {
	return nil
}

func (t *Teamvault) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate teamvault app ...")
	if err := t.Cluster.Validate(ctx); err != nil {
		return err
	}
	return nil
}
