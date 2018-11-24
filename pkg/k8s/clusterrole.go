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

func (d *ClusterRoleConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Context,
		d.ClusterRole,
	)
}

func (d *ClusterRoleConfiguration) Applier() (world.Applier, error) {
	return &ClusterRoleApplier{
		Context:     d.Context,
		ClusterRole: d.ClusterRole,
	}, nil
}

func (d *ClusterRoleConfiguration) Children() []world.Configuration {
	return d.Requirements
}

type ClusterRoleName string

func (n ClusterRoleName) String() string {
	return string(n)
}

func (a ClusterRoleName) Validate(ctx context.Context) error {
	if a == "" {
		return errors.New("ClusterRoleName missing")
	}
	return nil
}

type ClusterRoleApplier struct {
	Context     Context
	ClusterRole ClusterRole
}

func (s *ClusterRoleApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *ClusterRoleApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.ClusterRole,
	}
	return deployer.Apply(ctx)
}

func (s *ClusterRoleApplier) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.ClusterRole.Validate(ctx)
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

func (s ClusterRole) Validate(ctx context.Context) error {
	if s.ApiVersion != "rbac.authorization.k8s.io/v1" {
		return errors.New("invalid ApiVersion")
	}
	if s.Kind != "ClusterRole" {
		return errors.New("invalid Kind")
	}
	return validation.Validate(ctx,
		s.ApiVersion,
		s.Kind,
	)
}

func (n ClusterRole) String() string {
	return fmt.Sprintf("%s/%s", n.Kind, n.Metadata.Name)
}
