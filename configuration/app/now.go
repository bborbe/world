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

type Now struct {
	Context      k8s.Context
	Domains      k8s.IngressHosts
	Tag          docker.Tag
	Requirements []world.Configuration
}

func (n *Now) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		n.Context,
		n.Domains,
		n.Tag,
	)
}

func (n *Now) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, n.Requirements...)
	result = append(result, n.now()...)
	return result
}

func (n *Now) now() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/now",
		Tag:        n.Tag,
	}
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: n.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "now",
					Name:      "now",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:   n.Context,
			Namespace: "now",
			Name:      "now",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "now",
					Image: image,
					Requirement: &build.Now{
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
					Args: []k8s.Arg{"-logtostderr", "-v=2"},
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: "8080",
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
		},
		&deployer.ServiceDeployer{
			Context:   n.Context,
			Namespace: "now",
			Name:      "now",
			Ports:     []deployer.Port{port},
		},
		k8s.BuildIngressConfigurationWithCertManager(
			n.Context,
			"now",
			"now",
			"now",
			"http",
			"/",
			n.Domains...,
		),
	}
}

func (n *Now) Applier() (world.Applier, error) {
	return nil, nil
}
