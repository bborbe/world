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

type RoleBindingConfiguration struct {
	Context      Context
	RoleBinding  RoleBinding
	Requirements []world.Configuration
}

func (r *RoleBindingConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		r.Context,
		r.RoleBinding,
	)
}

func (r *RoleBindingConfiguration) Applier() (world.Applier, error) {
	return &RoleBindingApplier{
		Context:     r.Context,
		RoleBinding: r.RoleBinding,
	}, nil
}

func (r *RoleBindingConfiguration) Children() []world.Configuration {
	return r.Requirements
}

type RoleBindingName string

func (n RoleBindingName) String() string {
	return string(n)
}

func (a RoleBindingName) Validate(ctx context.Context) error {
	if a == "" {
		return errors.New("RoleBindingName missing")
	}
	return nil
}

type RoleBindingApplier struct {
	Context     Context
	RoleBinding RoleBinding
}

func (s *RoleBindingApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *RoleBindingApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.RoleBinding,
	}
	return deployer.Apply(ctx)
}

func (s *RoleBindingApplier) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.RoleBinding.Validate(ctx)
}

type RoleBinding struct {
	ApiVersion ApiVersion `yaml:"apiVersion"`
	Kind       Kind       `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
	Subjects   []Subject  `yaml:"subjects"`
	RoleRef    RoleRef    `yaml:"roleRef"`
}

func (r RoleBinding) Validate(ctx context.Context) error {
	if r.ApiVersion != "rbac.authorization.k8s.io/v1" {
		return errors.New("invalid ApiVersion")
	}
	if r.Kind != "RoleBinding" {
		return errors.New("invalid Kind")
	}
	return validation.Validate(ctx,
		r.ApiVersion,
		r.Kind,
	)
}

func (r RoleBinding) String() string {
	return fmt.Sprintf("%s/%s", r.Kind, r.Metadata.Name)
}
