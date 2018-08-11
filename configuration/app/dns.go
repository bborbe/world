package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/golang/glog"
)

type Dns struct {
	Cluster cluster.Cluster
}

func (d *Dns) Childs() []world.Configuration {
	return []world.Configuration{}
}

func (d *Dns) Applier() world.Applier {
	return nil
}

func (d *Dns) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate dns app ...")
	if err := d.Cluster.Validate(ctx); err != nil {
		return err
	}
	return nil
}
