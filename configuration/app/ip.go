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
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Ip struct {
	Context      k8s.Context
	Domains      k8s.IngressHosts
	Tag          docker.Tag
	IP           network.IP
	Requirements []world.Configuration
}

func (i *Ip) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		i.Context,
		i.Domains,
		i.Tag,
		i.IP,
	)
}

func (i *Ip) Applier() (world.Applier, error) {
	return nil, nil
}

func (i *Ip) Children() []world.Configuration {
	var result []world.Configuration
	result = append(result, i.Requirements...)
	result = append(result, i.ip()...)
	return result
}

func (i *Ip) ip() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/ip",
		Tag:        i.Tag,
	}
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: i.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "ip",
					Name:      "ip",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:   i.Context,
			Namespace: "ip",
			Name:      "ip",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "ip",
					Image: image,
					Requirement: &build.Ip{
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
					Args:  []k8s.Arg{"-logtostderr", "-v=2"},
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
			Context:   i.Context,
			Namespace: "ip",
			Name:      "ip",
			Ports:     []deployer.Port{port},
		},
		k8s.BuildIngressConfigurationWithCertManager(
			i.Context,
			"ip",
			"ip",
			"ip",
			"http",
			"/",
			i.Domains...,
		),
	}
}
