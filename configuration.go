package world

import (
	"context"
)

//go:generate counterfeiter -o mocks/configuration.go --fake-name Configuration . Configuration
type Configuration interface {
	Children() []Configuration
	Applier() Applier
	Validate(ctx context.Context) error
}

type ConfigurationBuilder struct {
	applier  Applier
	children []Configuration
	validate func(ctx context.Context) error
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

func (c *ConfigurationBuilder) WithValidate(validate func(ctx context.Context) error) *ConfigurationBuilder {
	c.validate = validate
	return c
}

func (c *ConfigurationBuilder) Children() []Configuration {
	return c.children
}

func (c *ConfigurationBuilder) Applier() Applier {
	return c.applier
}

func (c *ConfigurationBuilder) Validate(ctx context.Context) error {
	if c.validate != nil {
		return c.validate(ctx)
	}
	return nil
}
