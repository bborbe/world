package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Traefik struct {
	Cluster cluster.Cluster
}

func (t *Traefik) Childs() []world.Configuration {
	return []world.Configuration{}
}

func (t *Traefik) Applier() world.Applier {
	return nil
}

func (t *Traefik) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate traefik app ...")
	if err := t.Cluster.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate traefik app failed")
	}
	return nil
}
