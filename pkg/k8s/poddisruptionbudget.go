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

type PodDisruptionBudgetConfiguration struct {
	Context             Context
	PodDisruptionBudget PodDisruptionBudget
	Requirements        []world.Configuration
}

func (d *PodDisruptionBudgetConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Context,
		d.PodDisruptionBudget,
	)
}

func (d *PodDisruptionBudgetConfiguration) Applier() (world.Applier, error) {
	return &PodDisruptionBudgetApplier{
		Context:             d.Context,
		PodDisruptionBudget: d.PodDisruptionBudget,
	}, nil
}

func (d *PodDisruptionBudgetConfiguration) Children() []world.Configuration {
	return d.Requirements
}

type PodDisruptionBudgetApplier struct {
	Context             Context
	PodDisruptionBudget PodDisruptionBudget
}

func (s *PodDisruptionBudgetApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *PodDisruptionBudgetApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.PodDisruptionBudget,
	}
	return deployer.Apply(ctx)
}

func (s *PodDisruptionBudgetApplier) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.Context,
		s.PodDisruptionBudget,
	)
}

type PodDisruptionBudget struct {
	ApiVersion ApiVersion              `yaml:"apiVersion"`
	Kind       Kind                    `yaml:"kind"`
	Metadata   Metadata                `yaml:"metadata"`
	Spec       PodDisruptionBudgetSpec `yaml:"spec"`
}

func (s PodDisruptionBudget) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
}

func (s PodDisruptionBudget) Validate(ctx context.Context) error {
	if s.ApiVersion != "policy/v1beta1" {
		return errors.New("invalid ApiVersion")
	}
	if s.Kind != "PodDisruptionBudget" {
		return errors.New("invalid Kind")
	}
	return nil
}

type PodDisruptionBudgetSpec struct {
	MaxUnavailable int           `yaml:"maxUnavailable,omitempty"`
	MinAvailable   int           `yaml:"minAvailable,omitempty"`
	Selector       LabelSelector `yaml:"selector"`
}
