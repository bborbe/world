package deployer

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
)

type Port struct {
	Name     string
	Port     int
	HostPort int
	Protocol string
}

type DeploymentDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.Name
	Containers   []DeploymentDeployerContainer
	Volumes      []k8s.PodVolume
	HostNetwork  k8s.PodHostNetwork
	Requirements []world.Configuration
	DnsPolicy    k8s.PodDnsPolicy
	Labels       k8s.Labels
}

type DeploymentDeployerContainer struct {
	Name        k8s.PodName
	Args        []k8s.Arg
	Ports       []Port
	Env         []k8s.Env
	Resources   k8s.PodResources
	Mounts      []k8s.VolumeMount
	Image       docker.Image
	Requirement world.Configuration
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
		result = append(result, container.Requirement)
	}
	return result
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
			Selector: k8s.DeploymentSelector{
				MatchLabels: k8s.DeploymentMatchLabels{
					"app": d.Name.String(),
				},
			},
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Template: k8s.DeploymentTemplate{
				Metadata: k8s.Metadata{
					Labels: k8s.Labels{
						"app": d.Name.String(),
					},
				},
				Spec: k8s.PodSpec{
					Containers:  d.containers(),
					Volumes:     d.Volumes,
					HostNetwork: d.HostNetwork,
				},
			},
		},
	}
}

func (d *DeploymentDeployer) containers() []k8s.PodContainer {
	var result []k8s.PodContainer
	for _, container := range d.Containers {
		result = append(result, d.container(container))
	}
	return result
}

func (d *DeploymentDeployer) container(container DeploymentDeployerContainer) k8s.PodContainer {
	podContainer := k8s.PodContainer{
		Image:        k8s.PodImage(container.Image.String()),
		Name:         container.Name,
		Resources:    container.Resources,
		VolumeMounts: container.Mounts,
	}
	for _, port := range container.Ports {
		podContainer.Ports = append(podContainer.Ports, k8s.PodPort{
			Name:          k8s.PodPortName(port.Name),
			ContainerPort: k8s.PodPortContainerPort(port.Port),
			HostPort:      k8s.PodPortHostPort(port.HostPort),
			Protocol:      k8s.PodPortProtocol(port.Protocol),
		})
	}
	for _, arg := range container.Args {
		podContainer.Args = append(podContainer.Args, k8s.Arg(arg))
	}
	podContainer.Env = container.Env
	return podContainer
}
