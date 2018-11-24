// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
	"github.com/pkg/errors"
)

type DeploymentConfiguration struct {
	Context      Context
	Deployment   Deployment
	Requirements []world.Configuration
}

func (d *DeploymentConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Context,
		d.Deployment,
	)
}

func (d *DeploymentConfiguration) Applier() (world.Applier, error) {
	return &DeploymentApplier{
		Context:    d.Context,
		Deployment: d.Deployment,
	}, nil
}

func (d *DeploymentConfiguration) Children() []world.Configuration {
	return d.Requirements
}

type DeploymentApplier struct {
	Context    Context
	Deployment Deployment
}

func (s *DeploymentApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *DeploymentApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.Deployment,
	}
	return deployer.Apply(ctx)
}

func (s *DeploymentApplier) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.Context,
		s.Deployment,
	)
}

type Deployment struct {
	ApiVersion ApiVersion     `yaml:"apiVersion"`
	Kind       Kind           `yaml:"kind"`
	Metadata   Metadata       `yaml:"metadata"`
	Spec       DeploymentSpec `yaml:"spec"`
}

func (s Deployment) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
}

func (s Deployment) Validate(ctx context.Context) error {
	if s.ApiVersion != "apps/v1" {
		return errors.New("invalid ApiVersion")
	}
	if s.Kind != "Deployment" {
		return errors.New("invalid Kind")
	}
	return validation.Validate(
		ctx,
		s.Spec,
	)
}

type Replicas int

func (r Replicas) Int() int {
	return int(r)
}

func (r Replicas) String() string {
	return strconv.Itoa(r.Int())
}

func (s Replicas) Validate(ctx context.Context) error {
	if s.Int() <= 0 {
		return errors.New("invalid Relicas amount")
	}
	return nil
}

type DeploymentRevisionHistoryLimit int

type DeploymentSpec struct {
	Replicas             Replicas                       `yaml:"replicas"`
	RevisionHistoryLimit DeploymentRevisionHistoryLimit `yaml:"revisionHistoryLimit"`
	Selector             LabelSelector                  `yaml:"selector,omitempty"`
	Strategy             DeploymentStrategy             `yaml:"strategy"`
	Template             PodTemplate                    `yaml:"template"`
}

func (d DeploymentSpec) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Template,
		d.Selector,
	)
}

type LabelSelector struct {
	MatchLabels Labels `yaml:"matchLabels,omitempty"`
}

func (d LabelSelector) Validate(ctx context.Context) error {
	if len(d.MatchLabels) == 0 {
		return fmt.Errorf("MatchLabels empty")
	}
	return nil
}

type DeploymentMaxSurge int

type DeploymentMaxUnavailable int

type DeploymentStrategyType string

type DeploymentStrategyRollingUpdate struct {
	MaxSurge       DeploymentMaxSurge       `yaml:"maxSurge,omitempty"`
	MaxUnavailable DeploymentMaxUnavailable `yaml:"maxUnavailable,omitempty"`
}

type DeploymentStrategy struct {
	Type          DeploymentStrategyType          `yaml:"type"`
	RollingUpdate DeploymentStrategyRollingUpdate `yaml:"rollingUpdate,omitempty"`
}
