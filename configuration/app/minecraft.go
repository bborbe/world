// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"time"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Minecraft struct {
	Context k8s.Context
}

func (m *Minecraft) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		m.Context,
	)
}

func (m *Minecraft) Applier() (world.Applier, error) {
	return nil, nil
}

func (m *Minecraft) Children() []world.Configuration {
	version := "master"
	image := docker.Image{
		Repository: "bborbe/minecraft",
		Tag:        docker.TagWithTime(version, time.Now()),
	}
	serverPort := deployer.Port{
		Port:     25565,
		HostPort: 25565,
		Name:     "server",
		Protocol: "TCP",
	}
	rconPort := deployer.Port{
		Port:     25575,
		HostPort: 25575,
		Name:     "rcon",
		Protocol: "TCP",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: m.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "minecraft",
					Name:      "minecraft",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:   m.Context,
			Namespace: "minecraft",
			Name:      "minecraft",
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name: "server",
					Env: []k8s.Env{
						{
							Name:  "EULA",
							Value: "true",
						},
					},
					Image: image,
					Requirement: &build.Minecraft{
						Image: image,
					},
					Ports: []deployer.Port{
						serverPort,
						rconPort,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "2000m",
							Memory: "2000Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "500m",
							Memory: "500Mi",
						},
					},
					LivenessProbe: k8s.Probe{
						Exec: k8s.Exec{
							Command: []k8s.Command{
								"mcstatus",
								"localhost",
								"ping",
							},
						},
						InitialDelaySeconds: 240,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
					},
					ReadinessProbe: k8s.Probe{
						Exec: k8s.Exec{
							Command: []k8s.Command{
								"mcstatus",
								"localhost",
								"ping",
							},
						},
						InitialDelaySeconds: 120,
						TimeoutSeconds:      5,
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "datadir",
							Path: "/data",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "datadir",
					Host: k8s.PodVolumeHost{
						Path: "/data/minecraft",
					},
				},
			},
		},
	}
}
