package server

import (
	"context"

	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Nuke struct {
}

func (r *Nuke) Children() []world.Configuration {
	return nil
}

func (r *Nuke) Applier() (world.Applier, error) {
	return nil, nil
}

func (r *Nuke) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
	)
}
