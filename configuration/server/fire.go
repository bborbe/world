package server

import (
	"context"

	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Fire struct {
}

func (r *Fire) Children() []world.Configuration {
	return nil
}

func (r *Fire) Applier() (world.Applier, error) {
	return nil, nil
}

func (r *Fire) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
	)
}
