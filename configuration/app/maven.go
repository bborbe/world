package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Maven struct {
	Cluster          cluster.Cluster
	Domains          k8s.IngressHosts
	MavenRepoVersion docker.Tag
}

func (t *Maven) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
		t.Domains,
		t.MavenRepoVersion,
	)
}

func (m *Maven) Children() []world.Configuration {
	result := []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
		},
	}
	result = append(result, m.public()...)
	result = append(result, m.api()...)
	return result
}

func (m *Maven) public() []world.Configuration {
	port := deployer.Port{
		Port:     80,
		Name:     "http",
		Protocol: "TCP",
	}
	image := docker.Image{
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "public",
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
					Image: image,
					Requirement: &build.NginxAutoindex{
						Image: image,
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
							Name:     "maven",
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
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "maven",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/maven",
						Server: m.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "public",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "public",
			Port:      "http",
			Domains:   m.Domains,
		},
	}
}

func (m *Maven) api() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/maven-repo",
		Tag:        m.MavenRepoVersion,
	}
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "api",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "maven",
					Image: image,
					Requirement: &build.Maven{
						Image: image,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Args: []k8s.Arg{"-logtostderr", "-v=1"},
					Env: []k8s.Env{
						{
							Name:  "ROOT",
							Value: "/data",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "maven",
							Path: "/data",
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
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "maven",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/maven",
						Server: m.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Cluster.Context,
			Namespace: "maven",
			Name:      "api",
			Ports:     []deployer.Port{port},
		},
	}
}

func (m *Maven) Applier() (world.Applier, error) {
	return nil, nil
}
