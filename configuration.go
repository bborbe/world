package world

import "context"

//go:generate counterfeiter -o mocks/configuration.go --fake-name Configuration . Configuration
type Configuration interface {
	Children() []Configuration
	Applier() (Applier, error)
	Validate(ctx context.Context) error
}

type ConfigurationBuilder struct {
	applier  Applier
	children []Configuration
}

func NewConfiguration() *ConfigurationBuilder {
	return new(ConfigurationBuilder)
}

func (c *ConfigurationBuilder) WithChildren(children []Configuration) *ConfigurationBuilder {
	c.children = children
	return c
}

func (c *ConfigurationBuilder) WithApplier(applier Applier) *ConfigurationBuilder {
	c.applier = applier
	return c
}

func (c *ConfigurationBuilder) Children() []Configuration {
	return c.children
}

func (c *ConfigurationBuilder) Applier() (Applier, error) {
	return c.applier, nil
}

func (c *ConfigurationBuilder) Validate(ctx context.Context) error {
	return nil
}
