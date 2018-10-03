package world

import (
	"context"
)

type ConfiguraionBuilder struct {
	children []Configuration
	applier  Applier
}

func NewConfiguraionBuilder() *ConfiguraionBuilder {
	return &ConfiguraionBuilder{}
}

func (c *ConfiguraionBuilder) Children() []Configuration {
	return c.children
}

func (c *ConfiguraionBuilder) WithChildren(children []Configuration) *ConfiguraionBuilder {
	c.children = children
	return c
}

func (c *ConfiguraionBuilder) AddChildren(children ...Configuration) *ConfiguraionBuilder {
	c.children = append(c.children, children...)
	return c
}

func (c *ConfiguraionBuilder) Applier() (Applier, error) {
	return c.applier, nil
}

func (c *ConfiguraionBuilder) WithApplier(applier Applier) *ConfiguraionBuilder {
	c.applier = applier
	return c
}

func (c *ConfiguraionBuilder) WithApplierBuildFunc(applierBuildFunc ApplierBuildFunc) *ConfiguraionBuilder {
	c.applier = &ApplierBuilder{
		Build: applierBuildFunc,
	}
	return c
}

func (c *ConfiguraionBuilder) Validate(ctx context.Context) error {
	return nil
}
