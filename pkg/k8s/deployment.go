package k8s

import "fmt"

type Deployment struct {
	ApiVersion ApiVersion     `yaml:"apiVersion"`
	Kind       Kind           `yaml:"kind"`
	Metadata   Metadata       `yaml:"metadata"`
	Spec       DeploymentSpec `yaml:"spec"`
}

func (s Deployment) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
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
