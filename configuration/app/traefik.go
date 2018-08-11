package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/golang/glog"
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
		return err
	}
	return nil
}
