package server

import (
	"context"

	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Netcup struct {
}

func (r *Netcup) Children() []world.Configuration {
	return nil
}

func (r *Netcup) Applier() (world.Applier, error) {
	return nil, nil
}

func (r *Netcup) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
	)
}
