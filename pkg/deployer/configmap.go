// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deployer

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type ConfigValues map[string]ConfigValue

func (c ConfigValues) Checksum() string {
	hasher := md5.New()
	for _, value := range c {
		value, _ := value.Value(context.Background())
		hasher.Write([]byte(value))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (c ConfigValues) Validate(ctx context.Context) error {
	for _, value := range c {
		if err := value.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type ConfigValue interface {
	Value(ctx context.Context) (string, error)
	Validate(ctx context.Context) error
}

type ConfigValueStatic string

func (s ConfigValueStatic) Value(ctx context.Context) (string, error) {
	return string(s), nil
}

func (s ConfigValueStatic) Validate(ctx context.Context) error {
	return nil
}

type ConfigValueFunc func(ctx context.Context) (string, error)

func (s ConfigValueFunc) Value(ctx context.Context) (string, error) {
	return s(ctx)
}

func (s ConfigValueFunc) Validate(ctx context.Context) error {
	_, err := s(ctx)
	return err
}

type ConfigMapApplier struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.MetadataName
	ConfigValues ConfigValues
}

func (c *ConfigMapApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (c *ConfigMapApplier) Apply(ctx context.Context) error {
	configmap, err := c.configmap(ctx)
	if err != nil {
		return err
	}
	applier := &k8s.ConfigMapApplier{
		Context:   c.Context,
		ConfigMap: *configmap,
	}
	return applier.Apply(ctx)
}

func (c *ConfigMapApplier) configmap(ctx context.Context) (*k8s.ConfigMap, error) {
	data := make(k8s.ConfigMapData)
	for k, v := range c.ConfigValues {
		value, err := v.Value(ctx)
		if err != nil {
			return nil, err
		}
		data[k] = value
	}
	return &k8s.ConfigMap{
		ApiVersion: "v1",
		Kind:       "ConfigMap",
		Metadata: k8s.Metadata{
			Namespace: c.Namespace,
			Name:      c.Name,
			Labels: k8s.Labels{
				"app": c.Namespace.String(),
			},
		},
		Data: data,
	}, nil
}

func (c *ConfigMapApplier) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Context,
		c.Name,
		c.Namespace,
		c.ConfigValues,
	)
}
