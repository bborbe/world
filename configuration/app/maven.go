// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Maven struct {
	Context          k8s.Context
	Domains          k8s.IngressHosts
	MavenRepoVersion docker.Tag
	Requirements     []world.Configuration
}

func (m *Maven) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		m.Context,
		m.Domains,
		m.MavenRepoVersion,
	)
}

func (m *Maven) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, m.Requirements...)
	result = append(result, m.maven()...)
	return result
}

func (m *Maven) maven() []world.Configuration {
	result := []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: m.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "maven",
					Name:      "maven",
				},
			},
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
			Context:   m.Context,
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
					Host: k8s.PodVolumeHost{
						Path: "/data/maven",
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Context,
			Namespace: "maven",
			Name:      "public",
			Ports:     []deployer.Port{port},
		},
		k8s.BuildIngressConfigurationWithCertManager(
			m.Context,
			"maven",
			"maven",
			"public",
			"http",
			"/",
			m.Domains...,
		),
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
			Context:   m.Context,
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
					Host: k8s.PodVolumeHost{
						Path: "/data/maven",
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   m.Context,
			Namespace: "maven",
			Name:      "api",
			Ports:     []deployer.Port{port},
		},
	}
}

func (m *Maven) Applier() (world.Applier, error) {
	return nil, nil
}
