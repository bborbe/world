package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/pkg/validation"
)

type Prometheus struct {
	Cluster cluster.Cluster
}

func (t *Prometheus) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
	)
}

func (p *Prometheus) Children() []world.Configuration {
	return []world.Configuration{}
}

func (p *Prometheus) Applier() (world.Applier, error) {
	return nil, nil
}
