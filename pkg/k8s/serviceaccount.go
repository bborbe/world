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

type ServiceaccountConfiguration struct {
	Context        Context
	Serviceaccount ServiceAccount
	Requirements   []world.Configuration
}

func (d *ServiceaccountConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Context,
		d.Serviceaccount,
	)
}

func (d *ServiceaccountConfiguration) Applier() (world.Applier, error) {
	return &ServiceaccountApplier{
		Context:        d.Context,
		Serviceaccount: d.Serviceaccount,
	}, nil
}

func (d *ServiceaccountConfiguration) Children(ctx context.Context) (world.Configurations, error) {
	return d.Requirements, nil
}

type ServiceaccountApplier struct {
	Context        Context
	Serviceaccount ServiceAccount
}

func (s *ServiceaccountApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *ServiceaccountApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.Serviceaccount,
	}
	return deployer.Apply(ctx)
}

func (s *ServiceaccountApplier) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.Serviceaccount.Validate(ctx)
}

type ServiceAccount struct {
	ApiVersion ApiVersion `yaml:"apiVersion"`
	Kind       Kind       `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
}

func (s ServiceAccount) Validate(ctx context.Context) error {
	if s.ApiVersion != "v1" {
		return errors.New("invalid ApiVersion")
	}
	if s.Kind != "ServiceAccount" {
		return errors.New("invalid Kind")
	}
	return validation.Validate(ctx,
		s.ApiVersion,
		s.Kind,
		s.Metadata,
	)
}

func (n ServiceAccount) String() string {
	return fmt.Sprintf("%s/%s", n.Kind, n.Metadata.Name)
}
