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

type Debug struct {
	Context      k8s.Context
	Domains       k8s.IngressHosts
	Requirements []world.Configuration
}

func (d *Debug) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Context,
		d.Domains,
	)
}

func (d *Debug) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *Debug) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, d.Requirements...)
	result = append(result, d.debug()...)
	return result
}

func (d *Debug) debug() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/debug-server",
		Tag:        "1.0.0",
	}
	port := deployer.Port{
		Port:     8080,
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
					Namespace: "debug",
					Name:      "debug",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:   d.Context,
			Namespace: "debug",
			Name:      "debug",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name: "server",
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: port.Port.String(),
						},
					},
					Image: image,
					Requirement: &build.DebugServer{
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
		},
		&deployer.ServiceDeployer{
			Context:   d.Context,
			Namespace: "debug",
			Name:      "debug",
			Ports:     []deployer.Port{port},
		},
		k8s.BuildIngressConfigurationWithCertManager(
			d.Context,
			"debug",
			"debug",
			"debug",
			"http",
			"/",
			d.Domains...,
		),
	}
}
