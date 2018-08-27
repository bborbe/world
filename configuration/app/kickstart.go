package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Kickstart struct {
	Cluster cluster.Cluster
	Domains k8s.IngressHosts
}

func (t *Kickstart) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
	)
}

func (k *Kickstart) Applier() (world.Applier, error) {
	return nil, nil
}

func (k *Kickstart) Children() []world.Configuration {
	nginxImage := docker.Image{
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	port := deployer.Port{
		Port:     80,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kickstart",
		},
		&deployer.DeploymentDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kickstart",
			Name:      "kickstart",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "nginx",
					Image: nginxImage,
					Requirement: &build.NginxAutoindex{
						Image: nginxImage,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "250m",
							Memory: "25Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name:     "kickstart",
							Path:     "/usr/share/nginx/html",
							ReadOnly: true,
						},
					},
					Ports: []deployer.Port{port},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
					},
					ReadinessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
					},
				},
				&container.GitSync{
					MountName:  "kickstart",
					GitRepoUrl: "https://github.com/bborbe/kickstart.git",
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name:     "kickstart",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kickstart",
			Name:      "kickstart",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   k.Cluster.Context,
			Namespace: "kickstart",
			Name:      "kickstart",
			Port:      "http",
			Domains:   k.Domains,
		},
	}
}
