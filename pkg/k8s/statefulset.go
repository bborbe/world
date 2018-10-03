package k8s

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

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
	if s.ApiVersion != "apps/v1beta1" {
		return errors.New("invalid ApiVersion")
	}
	if s.Kind != "StatefulSet" {
		return errors.New("invalid Kind")
	}
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

type VolumeClaimTemplatesSpecResourcesRequests struct {
	Storage string `yaml:"storage,omitempty"`
}

type VolumeClaimTemplatesSpecResources struct {
	Requests VolumeClaimTemplatesSpecResourcesRequests `yaml:"requests,omitempty"`
}

type AccessMode string

func (a AccessMode) String() string {
	return string(a)
}

func (a AccessMode) Validate(ctx context.Context) error {
	if a == "" {
		return errors.New("invalid AccessMode")
	}
	return nil
}

type VolumeClaimTemplatesSpec struct {
	AccessModes []AccessMode                      `yaml:"accessModes,omitempty"`
	Resources   VolumeClaimTemplatesSpecResources `yaml:"resources,omitempty"`
}

type VolumeClaimTemplate struct {
	Metadata Metadata                 `yaml:"metadata,omitempty"`
	Spec     VolumeClaimTemplatesSpec `yaml:"spec,omitempty"`
}

type StatefulSetSpec struct {
	ServiceName          MetadataName          `yaml:"serviceName"`
	Replicas             Replicas              `yaml:"replicas"`
	Selector             Selector              `yaml:"selector,omitempty"`
	Template             PodTemplate           `yaml:"template"`
	VolumeClaimTemplates []VolumeClaimTemplate `yaml:"volumeClaimTemplates,omitempty"`
	UpdateStrategy       UpdateStrategy        `yaml:"updateStrategy,omitempty"`
}

func (s StatefulSetSpec) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.ServiceName,
		s.Replicas,
		s.Template,
	)
}

type UpdateStrategy struct {
	Type string `yaml:"type,omitempty"`
}
