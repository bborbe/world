package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Slideshow struct {
	Cluster cluster.Cluster
	Domains k8s.IngressHosts
}

func (t *Slideshow) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
	)
}

func (s *Slideshow) Applier() (world.Applier, error) {
	return nil, nil
}

func (s *Slideshow) Children() []world.Configuration {
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
			Context:   s.Cluster.Context,
			Namespace: "slideshow",
		},
		&deployer.DeploymentDeployer{
			Context:   s.Cluster.Context,
			Namespace: "slideshow",
			Name:      "slideshow",
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
					Ports: []deployer.Port{port},
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
							Name:     "slideshow",
							Path:     "/usr/share/nginx/html",
							ReadOnly: true,
						},
					},
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
					MountName:  "slideshow",
					GitRepoUrl: "https://github.com/bborbe/slideshow.git",
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name:     "slideshow",
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   s.Cluster.Context,
			Namespace: "slideshow",
			Name:      "slideshow",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   s.Cluster.Context,
			Namespace: "slideshow",
			Name:      "slideshow",
			Port:      "http",
			Domains:   s.Domains,
		},
	}
}