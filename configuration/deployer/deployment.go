package deployer

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
	"github.com/pkg/errors"
)

type Port struct {
	Name     k8s.PortName
	Port     k8s.PortNumber
	HostPort k8s.PortNumber
	Protocol k8s.PortProtocol
}

func (p *Port) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.Name,
		p.Port,
		p.Protocol,
	)
}

func (p Port) ContainerPort() k8s.ContainerPort {
	return k8s.ContainerPort{
		Name:          p.Name,
		ContainerPort: p.Port,
		HostPort:      p.HostPort,
		Protocol:      p.Protocol,
	}
}

func (p Port) ServicePort() k8s.ServicePort {
	return k8s.ServicePort{
		Name:     p.Name,
		Port:     p.Port,
		Protocol: p.Protocol,
	}
}

type HasContainer interface {
	Validate(ctx context.Context) error
	Container() k8s.Container
	Requirements() []world.Configuration
}

type DeploymentDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.MetadataName
	Containers   []HasContainer
	Volumes      []k8s.PodVolume
	HostNetwork  k8s.PodHostNetwork
	Requirements []world.Configuration
	DnsPolicy    k8s.PodDnsPolicy
	Labels       k8s.Labels
	Strategy     k8s.DeploymentStrategy
}

func (w *DeploymentDeployer) Validate(ctx context.Context) error {
	for _, container := range w.Containers {
		if err := container.Validate(ctx); err != nil {
			return errors.Wrap(err, "validate container failed")
		}
	}

	if w.Strategy.Type == "" {
		return errors.New("Deployment Strategy missing")
	}

	return validation.Validate(
		ctx,
		w.Context,
		w.Namespace,
		w.Name,
	)
}

type DeploymentDeployerContainer struct {
	Name           k8s.ContainerName
	Command        []k8s.Command
	Args           []k8s.Arg
	Ports          []Port
	Env            []k8s.Env
	Resources      k8s.Resources
	Mounts         []k8s.ContainerMount
	Image          docker.Image
	Requirement    world.Configuration
	LivenessProbe  k8s.Probe
	ReadinessProbe k8s.Probe
}

func (d *DeploymentDeployerContainer) Validate(ctx context.Context) error {
	return nil
}

func (d *DeploymentDeployer) Applier() (world.Applier, error) {
	return &k8s.DeploymentApplier{
		Context:    d.Context,
		Deployment: d.deployment(),
	}, nil
}

func (d *DeploymentDeployer) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, d.Requirements...)
	for _, container := range d.Containers {
		result = append(result, container.Requirements()...)
	}
	return result
}

func (d *DeploymentDeployer) WithRollingUpdate() *DeploymentDeployer {
	d.Strategy = k8s.DeploymentStrategy{
		Type: "RollingUpdate",
		RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
			MaxSurge:       1,
			MaxUnavailable: 1,
		},
	}
	return d
}

func (d *DeploymentDeployer) WithRecreate() *DeploymentDeployer {
	d.Strategy = k8s.DeploymentStrategy{
		Type: "Recreate",
	}
	return d
}

func (d *DeploymentDeployer) deployment() k8s.Deployment {
	return k8s.Deployment{
		ApiVersion: "extensions/v1beta1",
		Kind:       "Deployment",
		Metadata: k8s.Metadata{
			Namespace: d.Namespace,
			Name:      d.Name,
			Labels: k8s.Labels{
				"app": d.Name.String(),
			},
		},
		Spec: k8s.DeploymentSpec{
			Replicas:             1,
			RevisionHistoryLimit: 2,
			Selector: k8s.Selector{
				MatchLabels: k8s.Labels{
					"app": d.Name.String(),
				},
			},
			Strategy: d.Strategy,
			Template: k8s.PodTemplate{
				Metadata: k8s.Metadata{
					Labels: k8s.Labels{
						"app": d.Name.String(),
					},
				},
				Spec: k8s.PodSpec{
					Containers:  d.containers(),
					Volumes:     d.Volumes,
					HostNetwork: d.HostNetwork,
					DnsPolicy:   d.DnsPolicy,
				},
			},
		},
	}
}

func (d *DeploymentDeployer) containers() []k8s.Container {
	var result []k8s.Container
	for _, container := range d.Containers {
		result = append(result, container.Container())
	}
	return result
}

func (d *DeploymentDeployerContainer) Requirements() []world.Configuration {
	var result []world.Configuration
	if d.Requirement != nil {
		result = append(result, d.Requirement)
	}
	return result
}

func (d *DeploymentDeployerContainer) Container() k8s.Container {
	podContainer := k8s.Container{
		Image:          k8s.Image(d.Image.String()),
		Name:           d.Name,
		Resources:      d.Resources,
		VolumeMounts:   d.Mounts,
		LivenessProbe:  d.LivenessProbe,
		ReadinessProbe: d.ReadinessProbe,
		Args:           d.Args,
		Command:        d.Command,
		Env:            d.Env,
	}
	for _, port := range d.Ports {
		podContainer.Ports = append(podContainer.Ports, port.ContainerPort())
	}
	return podContainer
}
