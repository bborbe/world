package k8s

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/world"

	"github.com/bborbe/world/pkg/validation"
)

type StatefulSetConfiguration struct {
	Context      Context
	Requirements []world.Configuration
	StatefulSet  StatefulSet
}

func (w *StatefulSetConfiguration) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.Context,
	)
}

func (n *StatefulSetConfiguration) Applier() (world.Applier, error) {
	return &StatefulSetApplier{
		Context:     n.Context,
		StatefulSet: n.StatefulSet,
	}, nil
}

func (n *StatefulSetConfiguration) Children() []world.Configuration {
	return n.Requirements
}

type StatefulSetApplier struct {
	Context     Context
	StatefulSet StatefulSet
}

func (s *StatefulSetApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *StatefulSetApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.StatefulSet,
	}
	return deployer.Apply(ctx)
}

func (s *StatefulSetApplier) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.Context,
		s.StatefulSet,
	)
}

type StatefulSet struct {
	ApiVersion ApiVersion      `yaml:"apiVersion"`
	Kind       Kind            `yaml:"kind"`
	Metadata   Metadata        `yaml:"metadata"`
	Spec       StatefulSetSpec `yaml:"spec"`
}

func (s StatefulSet) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.ApiVersion,
		s.Kind,
		s.Metadata,
		s.Spec,
	)
}

func (s StatefulSet) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
}

type VolumeClaimTemplates struct {
	Metadata Metadata `yaml:"metadata,omitempty"`
}

type StatefulSetSpec struct {
	ServiceName          MetadataName         `yaml:"serviceName"`
	Replicas             Replicas             `yaml:"replicas"`
	Selector             Selector             `yaml:"selector,omitempty"`
	Template             PodTemplate          `yaml:"template"`
	VolumeClaimTemplates VolumeClaimTemplates `yaml:"volumeClaimTemplates,omitempty"`
}

func (s StatefulSetSpec) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.ServiceName,
		s.Replicas,
		s.Template,
	)
}
