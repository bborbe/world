package k8s

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

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
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.Deployment.Validate(ctx)
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

func (s *Deployment) Validate(ctx context.Context) error {
	return nil
}

type DeploymentReplicas int

type DeploymentRevisionHistoryLimit int

type DeploymentSpec struct {
	Replicas             DeploymentReplicas             `yaml:"replicas"`
	RevisionHistoryLimit DeploymentRevisionHistoryLimit `yaml:"revisionHistoryLimit"`
	Selector             DeploymentSelector             `yaml:"selector"`
	Strategy             DeploymentStrategy             `yaml:"strategy"`
	Template             DeploymentTemplate             `yaml:"template"`
}

type DeploymentMatchLabels map[string]string

type DeploymentSelector struct {
	MatchLabels DeploymentMatchLabels `yaml:"matchLabels,omitempty"`
}

type DeploymentMaxSurge int

type DeploymentMaxUnavailable int

type DeploymentStrategyType string

type DeploymentStrategyRollingUpdate struct {
	MaxSurge       DeploymentMaxSurge       `yaml:"maxSurge"`
	MaxUnavailable DeploymentMaxUnavailable `yaml:"maxUnavailable"`
}

type DeploymentStrategy struct {
	Type          DeploymentStrategyType          `yaml:"type"`
	RollingUpdate DeploymentStrategyRollingUpdate `yaml:"rollingUpdate"`
}

type DeploymentTemplate struct {
	Metadata Metadata `yaml:"metadata"`
	Spec     PodSpec  `yaml:"spec"`
}
