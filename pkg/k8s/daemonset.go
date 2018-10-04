package k8s

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/world"

	"github.com/bborbe/world/pkg/validation"
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

func (c DaemonSet) Validate(ctx context.Context) error {
	if c.ApiVersion != "apps/v1" {
		return errors.New("invalid ApiVersion")
	}
	if c.Kind != "DaemonSet" {
		return errors.New("invalid Kind")
	}
	return validation.Validate(
		ctx,
		c.ApiVersion,
		c.Kind,
		c.Metadata,
		c.Spec,
	)
}

func (s DaemonSet) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
}

type DaemonSetSpec struct {
	Selector LabelSelector `yaml:"selector,omitempty"`
	Template PodTemplate   `yaml:"template"`
}

func (c DaemonSetSpec) Validate(ctx context.Context) error {
	for key, value := range c.Selector.MatchLabels {
		v, ok := c.Template.Metadata.Labels[key]
		if !ok || v != value {
			return fmt.Errorf("label %s not in selector and not in metadata", key)
		}
	}
	return validation.Validate(
		ctx,
		c.Template,
	)
}
