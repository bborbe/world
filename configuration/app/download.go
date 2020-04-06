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

type Download struct {
	Context      k8s.Context
	Domains      k8s.IngressHosts
	Requirements []world.Configuration
}

func (d *Download) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Context,
		d.Domains,
	)
}

func (d *Download) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *Download) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, d.Requirements...)
	result = append(result, d.download()...)
	return result
}

func (d *Download) download() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/nginx-autoindex",
		Tag:        "latest",
	}
	port := deployer.Port{
		Port:     80,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: d.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "download",
					Name:      "download",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:   d.Context,
			Namespace: "download",
			Name:      "download",
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
							Name:     "download",
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
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "download",
					Host: k8s.PodVolumeHost{
						Path: "/data/download",
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   d.Context,
			Namespace: "download",
			Name:      "download",
			Ports:     []deployer.Port{port},
		},
		k8s.BuildIngressConfigurationWithCertManager(
			d.Context,
			"download",
			"download",
			"download",
			"http",
			"/",
			d.Domains...,
		),
	}
}
