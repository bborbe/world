// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type HelloWorld struct {
	Context k8s.Context
	Domains k8s.IngressHosts
	Tag     docker.Tag
}

func (h *HelloWorld) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		h.Context,
		h.Domains,
		h.Tag,
	)
}

func (h *HelloWorld) Applier() (world.Applier, error) {
	return nil, nil
}

func (h *HelloWorld) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/hello-world",
		Tag:        h.Tag,
	}
	port := deployer.Port{
		Port:     80,
		Name:     "http",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: h.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "hello-world",
					Name:      "hello-world",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:   h.Context,
			Namespace: "hello-world",
			Name:      "hello-world",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "hello-world",
					Image: image,
					Requirement: &build.HelloWorld{
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
			Context:   h.Context,
			Namespace: "hello-world",
			Name:      "hello-world",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   h.Context,
			Namespace: "hello-world",
			Name:      "hello-world",
			Port:      "http",
			Domains:   h.Domains,
		},
	}
}
