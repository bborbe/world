// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type ConfigMapConfiguration struct {
	Context      Context
	ConfigMap    ConfigMap
	Requirements []world.Configuration
}

func (d *ConfigMapConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Context,
		d.ConfigMap,
	)
}

func (d *ConfigMapConfiguration) Applier() (world.Applier, error) {
	return &ConfigMapApplier{
		Context:   d.Context,
		ConfigMap: d.ConfigMap,
	}, nil
}

func (d *ConfigMapConfiguration) Children(ctx context.Context) (world.Configurations, error) {
	return d.Requirements, nil
}

type ConfigMapApplier struct {
	Context   Context
	ConfigMap ConfigMap
}

func (s *ConfigMapApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *ConfigMapApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.ConfigMap,
	}
	return deployer.Apply(ctx)
}

func (s *ConfigMapApplier) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.ConfigMap.Validate(ctx)
}

type ConfigMap struct {
	ApiVersion ApiVersion    `yaml:"apiVersion"`
	Kind       Kind          `yaml:"kind"`
	Metadata   Metadata      `yaml:"metadata"`
	Data       ConfigMapData `yaml:"data"`
}

func (c ConfigMap) Validate(ctx context.Context) error {
	if c.ApiVersion != "v1" {
		return errors.New("invalid ApiVersion")
	}
	if c.Kind != "ConfigMap" {
		return errors.New("invalid Kind")
	}
	return nil
}

func (c ConfigMap) String() string {
	return fmt.Sprintf("%s/%s to %s", c.Kind, c.Metadata.Name, c.Metadata.Namespace)
}

type ConfigMapType string

type ConfigMapData map[string]string

func (d ConfigMapData) Validate(ctx context.Context) error {
	for k := range d {
		if k == "" {
			return errors.New("Config has no name")
		}
	}
	return nil
}
