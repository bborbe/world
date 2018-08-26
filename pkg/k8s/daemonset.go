package k8s

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/validation"
)

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
	Selector Selector    `yaml:"selector,omitempty"`
	Template PodTemplate `yaml:"template"`
}

func (c DaemonSetSpec) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Template,
	)
}
