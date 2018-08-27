package build

import (
	"context"

	"github.com/bborbe/world"
)

type buildConfiguration struct {
	applier world.Applier
}

func (c *buildConfiguration) Applier() (world.Applier, error) {
	return c.applier, nil
}

func (c *buildConfiguration) Children() []world.Configuration {
	return nil
}

func (c *buildConfiguration) Validate(ctx context.Context) error {
	return nil
}
