// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package world

import (
	"context"
)

type ConfiguraionBuilder struct {
	children Configurations
	applier  Applier
}

func NewConfiguraionBuilder() *ConfiguraionBuilder {
	return &ConfiguraionBuilder{}
}

func (c *ConfiguraionBuilder) Children(ctx context.Context) (Configurations, error) {
	return c.children, nil
}

func (c *ConfiguraionBuilder) WithChildren(children Configurations) *ConfiguraionBuilder {
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
