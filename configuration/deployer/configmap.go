// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deployer

import (
	"context"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type ConfigEntryList []ConfigEntry

func (c ConfigEntryList) Validate(ctx context.Context) error {
	for _, e := range c {
		if err := e.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c ConfigEntryList) ConfigMapData(ctx context.Context) (k8s.ConfigMapData, error) {
	result := make(k8s.ConfigMapData)
	for _, e := range c {
		if e.ValueFrom != nil {
			valueFrom, err := e.ValueFrom(ctx)
			if err != nil {
				return nil, err
			}
			result[e.Key] = valueFrom
		} else {
			result[e.Key] = e.Value
		}
	}
	return result, nil
}

type ConfigEntry struct {
	Key       string
	Value     string
	ValueFrom func(ctx context.Context) (string, error)
}

func (c ConfigEntry) Validate(ctx context.Context) error {
	if c.Value != "" {
		return nil
	}
	if c.ValueFrom != nil {
		return nil
	}
	return errors.New("value and valueFrom empty")
}

type ConfigMapApplier struct {
	Context         k8s.Context
	Namespace       k8s.NamespaceName
	Name            k8s.MetadataName
	ConfigEntryList ConfigEntryList
}

func (c *ConfigMapApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (c *ConfigMapApplier) Apply(ctx context.Context) error {
	configMapData, err := c.ConfigEntryList.ConfigMapData(ctx)
	if err != nil {
		return errors.Wrap(err, "generate configMapData failed")
	}
	applier := &k8s.ConfigMapApplier{
		Context: c.Context,
		ConfigMap: k8s.ConfigMap{
			ApiVersion: "v1",
			Kind:       "ConfigMap",
			Metadata: k8s.Metadata{
				Namespace: c.Namespace,
				Name:      c.Name,
				Labels: k8s.Labels{
					"app": c.Namespace.String(),
				},
			},
			Data: configMapData,
		},
	}
	return applier.Apply(ctx)
}

func (c *ConfigMapApplier) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Context,
		c.Name,
		c.Namespace,
		c.ConfigEntryList,
	)
}
