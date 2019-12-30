// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Ip struct {
	Context k8s.Context
	Domain  k8s.IngressHost
	Tag     docker.Tag
	IP      network.IP
}

func (t *Ip) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Context,
		t.Domain,
		t.Tag,
		t.IP,
	)
}

func (i *Ip) Applier() (world.Applier, error) {
	return nil, nil
}

func (i *Ip) Children() []world.Configuration {
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
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: dns.Host(i.Domain.String()),
						IP:   i.IP,
					},
				},
			},
		),
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
		&deployer.IngressDeployer{
			Context:   i.Context,
			Namespace: "ip",
			Name:      "ip",
			Port:      "http",
			Domains:   k8s.IngressHosts{i.Domain},
		},
	}
}
