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

type ClusterRoleBindingConfiguration struct {
	Context            Context
	ClusterRoleBinding ClusterRoleBinding
	Requirements       []world.Configuration
}

func (c *ClusterRoleBindingConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Context,
		c.ClusterRoleBinding,
	)
}

func (c *ClusterRoleBindingConfiguration) Applier() (world.Applier, error) {
	return &ClusterRoleBindingApplier{
		Context:            c.Context,
		ClusterRoleBinding: c.ClusterRoleBinding,
	}, nil
}

func (c *ClusterRoleBindingConfiguration) Children(ctx context.Context) (world.Configurations, error) {
	return c.Requirements, nil
}

type ClusterRoleBindingName string

func (n ClusterRoleBindingName) String() string {
	return string(n)
}

func (a ClusterRoleBindingName) Validate(ctx context.Context) error {
	if a == "" {
		return errors.New("ClusterRoleBindingName missing")
	}
	return nil
}

type ClusterRoleBindingApplier struct {
	Context            Context
	ClusterRoleBinding ClusterRoleBinding
}

func (s *ClusterRoleBindingApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *ClusterRoleBindingApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.ClusterRoleBinding,
	}
	return deployer.Apply(ctx)
}

func (s *ClusterRoleBindingApplier) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.ClusterRoleBinding.Validate(ctx)
}

type Subject struct {
	Kind      Kind   `yaml:"kind,omitempty"`
	Name      string `yaml:"name,omitempty"`
	ApiGroup  string `yaml:"apiGroup,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

type RoleRef struct {
	Kind     Kind   `yaml:"kind,omitempty"`
	Name     string `yaml:"name,omitempty"`
	ApiGroup string `yaml:"apiGroup,omitempty"`
}

type ClusterRoleBinding struct {
	ApiVersion ApiVersion `yaml:"apiVersion"`
	Kind       Kind       `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
	Subjects   []Subject  `yaml:"subjects"`
	RoleRef    RoleRef    `yaml:"roleRef"`
}

func (s ClusterRoleBinding) Validate(ctx context.Context) error {
	if s.ApiVersion != "rbac.authorization.k8s.io/v1" {
		return errors.New("invalid ApiVersion")
	}
	if s.Kind != "ClusterRoleBinding" {
		return errors.New("invalid Kind")
	}
	return validation.Validate(ctx,
		s.ApiVersion,
		s.Kind,
	)
}

func (n ClusterRoleBinding) String() string {
	return fmt.Sprintf("%s/%s", n.Kind, n.Metadata.Name)
}
