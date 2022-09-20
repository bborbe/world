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

type RoleConfiguration struct {
	Context      Context
	Role         Role
	Requirements []world.Configuration
}

func (r *RoleConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		r.Context,
		r.Role,
	)
}

func (r *RoleConfiguration) Applier() (world.Applier, error) {
	return &RoleApplier{
		Context: r.Context,
		Role:    r.Role,
	}, nil
}

func (r *RoleConfiguration) Children(ctx context.Context) (world.Configurations, error) {
	return r.Requirements, nil
}

type RoleName string

func (r RoleName) String() string {
	return string(r)
}

func (r RoleName) Validate(ctx context.Context) error {
	if r == "" {
		return errors.New("RoleName missing")
	}
	return nil
}

type RoleApplier struct {
	Context Context
	Role    Role
}

func (s *RoleApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *RoleApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.Role,
	}
	return deployer.Apply(ctx)
}

func (s *RoleApplier) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.Role.Validate(ctx)
}

type Role struct {
	ApiVersion ApiVersion   `yaml:"apiVersion,omitempty"`
	Kind       Kind         `yaml:"kind,omitempty"`
	Metadata   Metadata     `yaml:"metadata,omitempty"`
	Rules      []PolicyRule `yaml:"rules,omitempty"`
}

func (c Role) Validate(ctx context.Context) error {
	if c.ApiVersion != "rbac.authorization.k8s.io/v1" {
		return errors.New("invalid ApiVersion")
	}
	if c.Kind != "Role" {
		return errors.New("invalid Kind")
	}
	return validation.Validate(ctx,
		c.ApiVersion,
		c.Kind,
	)
}

func (c Role) String() string {
	return fmt.Sprintf("%s/%s", c.Kind, c.Metadata.Name)
}
