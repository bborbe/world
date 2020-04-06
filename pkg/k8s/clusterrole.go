// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
	"github.com/pkg/errors"
)

type ClusterRoleConfiguration struct {
	Context      Context
	ClusterRole  ClusterRole
	Requirements []world.Configuration
}

func (c *ClusterRoleConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Context,
		c.ClusterRole,
	)
}

func (c *ClusterRoleConfiguration) Applier() (world.Applier, error) {
	return &ClusterRoleApplier{
		Context:     c.Context,
		ClusterRole: c.ClusterRole,
	}, nil
}

func (c *ClusterRoleConfiguration) Children() []world.Configuration {
	return c.Requirements
}

type ClusterRoleName string

func (c ClusterRoleName) String() string {
	return string(c)
}

func (c ClusterRoleName) Validate(ctx context.Context) error {
	if c == "" {
		return errors.New("ClusterRoleName missing")
	}
	return nil
}

type ClusterRoleApplier struct {
	Context     Context
	ClusterRole ClusterRole
}

func (c *ClusterRoleApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (c *ClusterRoleApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: c.Context,
		Data:    c.ClusterRole,
	}
	return deployer.Apply(ctx)
}

func (c *ClusterRoleApplier) Validate(ctx context.Context) error {
	if c.Context == "" {
		return errors.New("context missing")
	}
	return c.ClusterRole.Validate(ctx)
}

type PolicyRule struct {
	ApiGroups       []string `yaml:"apiGroups,omitempty"`
	NonResourceURLs []string `yaml:"nonResourceURLs,omitempty"`
	ResourceNames   []string `yaml:"resourceNames,omitempty"`
	Resources       []string `yaml:"resources,omitempty"`
	Verbs           []string `yaml:"verbs,omitempty"`
}

type ClusterRole struct {
	ApiVersion ApiVersion   `yaml:"apiVersion,omitempty"`
	Kind       Kind         `yaml:"kind,omitempty"`
	Metadata   Metadata     `yaml:"metadata,omitempty"`
	Rules      []PolicyRule `yaml:"rules,omitempty"`
}

func (c ClusterRole) Validate(ctx context.Context) error {
	if c.ApiVersion != "rbac.authorization.k8s.io/v1" {
		return errors.New("invalid ApiVersion")
	}
	if c.Kind != "ClusterRole" {
		return errors.New("invalid Kind")
	}
	return validation.Validate(ctx,
		c.ApiVersion,
		c.Kind,
	)
}

func (c ClusterRole) String() string {
	return fmt.Sprintf("%s/%s", c.Kind, c.Metadata.Name)
}
