package configuration

import (
	"context"

	"github.com/bborbe/world/pkg/world"
)

type Builder struct {
	children []world.Configuration
	applier  world.Applier
}

func New() *Builder {
	return &Builder{}
}

func (c *Builder) Children() []world.Configuration {
	return c.children
}

func (c *Builder) WithChildren(children []world.Configuration) *Builder {
	c.children = children
	return c
}

func (c *Builder) AddChildren(children ...world.Configuration) *Builder {
	c.children = append(c.children, children...)
	return c
}

func (c *Builder) Applier() (world.Applier, error) {
	return c.applier, nil
}

func (c *Builder) WithApplier(applier world.Applier) *Builder {
	c.applier = applier
	return c
}

func (c *Builder) Validate(ctx context.Context) error {
	return nil
}
