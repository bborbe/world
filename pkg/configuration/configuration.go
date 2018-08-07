package configuration

import (
	"context"

	"github.com/bborbe/world"
)

type Configuration struct {
	applier  world.Applier
	childs   []world.Configuration
	validate func(ctx context.Context) error
}

func New() *Configuration {
	return new(Configuration)
}

func (c *Configuration) WithChilds(childs []world.Configuration) *Configuration {
	c.childs = childs
	return c
}

func (c *Configuration) WithApplier(applier world.Applier) *Configuration {
	c.applier = applier
	return c
}

func (c *Configuration) WithValidate(validate func(ctx context.Context) error) *Configuration {
	c.validate = validate
	return c
}

func (c *Configuration) Childs() []world.Configuration {
	return c.childs
}

func (c *Configuration) Applier() world.Applier {
	return c.applier
}

func (c *Configuration) Validate(ctx context.Context) error {
	if c.validate != nil {
		return c.validate(ctx)
	}
	return nil
}
