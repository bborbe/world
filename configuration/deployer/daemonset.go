package deployer

import (
	"context"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type DaemonSetDeployer struct {
	Context      k8s.Context
	DaemonSet    k8s.DaemonSet
	Requirements []world.Configuration
}

func (d *DaemonSetDeployer) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Context,
		d.DaemonSet,
	)
}

func (d *DaemonSetDeployer) Applier() (world.Applier, error) {
	return &k8s.DaemonSetApplier{
		Context:   d.Context,
		DaemonSet: d.DaemonSet,
	}, nil
}

func (d *DaemonSetDeployer) Children() []world.Configuration {
	return d.Requirements
}
