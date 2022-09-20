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

type StorageClassConfiguration struct {
	Context      Context
	StorageClass StorageClass
	Requirements []world.Configuration
}

func (d *StorageClassConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Context,
		d.StorageClass,
	)
}

func (d *StorageClassConfiguration) Applier() (world.Applier, error) {
	return &StorageClassApplier{
		Context:      d.Context,
		StorageClass: d.StorageClass,
	}, nil
}

func (d *StorageClassConfiguration) Children(ctx context.Context) (world.Configurations, error) {
	return d.Requirements, nil
}

type StorageClassApplier struct {
	Context      Context
	StorageClass StorageClass
}

func (s *StorageClassApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *StorageClassApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.StorageClass,
	}
	return deployer.Apply(ctx)
}

func (s *StorageClassApplier) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.StorageClass.Validate(ctx)
}

type StorageClassProvisioner string

type StorageClassParameters map[string]string

type StorageClass struct {
	ApiVersion  ApiVersion              `yaml:"apiVersion"`
	Kind        Kind                    `yaml:"kind"`
	Metadata    Metadata                `yaml:"metadata"`
	Provisioner StorageClassProvisioner `yaml:"provisioner"`
	Parameters  StorageClassParameters  `yaml:"parameters,omitempty"`
}

func (s StorageClass) Validate(ctx context.Context) error {
	if s.ApiVersion != "storage.k8s.io/v1" {
		return errors.New("invalid ApiVersion")
	}
	if s.Kind != "StorageClass" {
		return errors.New("invalid Kind")
	}
	return validation.Validate(ctx,
		s.ApiVersion,
		s.Kind,
		s.Metadata,
	)
}

func (n StorageClass) String() string {
	return fmt.Sprintf("%s/%s", n.Kind, n.Metadata.Name)
}

type StorageClassName string

func (s StorageClassName) String() string {
	return string(s)
}

func (s StorageClassName) Validate(ctx context.Context) error {
	if s == "" {
		return errors.New("invalid StorageClassName")
	}
	return nil
}
