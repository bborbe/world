package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Dns struct {
	Cluster cluster.Cluster
}

func (d *Dns) Children() []world.Configuration {
	return []world.Configuration{}
}

func (d *Dns) Applier() world.Applier {
	return nil
}

func (d *Dns) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate dns app ...")
	if err := d.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate dns app failed")
	}
	return nil
}
