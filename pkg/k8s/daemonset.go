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

type DaemonSetConfiguration struct {
	Context      Context
	DaemonSet    DaemonSet
	Requirements []world.Configuration
}

func (d *DaemonSetConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Context,
		d.DaemonSet,
	)
}

func (d *DaemonSetConfiguration) Applier() (world.Applier, error) {
	return &DaemonSetApplier{
		Context:   d.Context,
		DaemonSet: d.DaemonSet,
	}, nil
}

func (d *DaemonSetConfiguration) Children() []world.Configuration {
	return d.Requirements
}

type DaemonSetApplier struct {
	Context   Context
	DaemonSet DaemonSet
}

func (s *DaemonSetApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *DaemonSetApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.DaemonSet,
	}
	return deployer.Apply(ctx)
}

func (s *DaemonSetApplier) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.Context,
		s.DaemonSet,
	)
}

type DaemonSet struct {
	ApiVersion ApiVersion    `yaml:"apiVersion"`
	Kind       Kind          `yaml:"kind"`
	Metadata   Metadata      `yaml:"metadata"`
	Spec       DaemonSetSpec `yaml:"spec"`
}

func (d DaemonSet) Validate(ctx context.Context) error {
	if d.ApiVersion != "apps/v1" {
		return errors.New("invalid ApiVersion")
	}
	if d.Kind != "DaemonSet" {
		return errors.New("invalid Kind")
	}
	return validation.Validate(
		ctx,
		d.ApiVersion,
		d.Kind,
		d.Metadata,
		d.Spec,
	)
}

func (s DaemonSet) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
}

type DaemonSetSpec struct {
	Selector LabelSelector `yaml:"selector,omitempty"`
	Template PodTemplate   `yaml:"template"`
}

func (d DaemonSetSpec) Validate(ctx context.Context) error {
	for key, value := range d.Selector.MatchLabels {
		v, ok := d.Template.Metadata.Labels[key]
		if !ok || v != value {
			return fmt.Errorf("label %s not in selector and not in metadata", key)
		}
	}
	return validation.Validate(
		ctx,
		d.Template,
		d.Selector,
	)
}
