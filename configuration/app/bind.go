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

type Bind struct {
	Context k8s.Context
}

func (b *Bind) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		b.Context,
	)
}

func (b *Bind) Children() []world.Configuration {
	version := "1.1.0"
	image := docker.Image{
		Repository: "bborbe/bind",
		Tag:        docker.TagWithTime(version, time.Now()),
	}
	udpPort := deployer.Port{
		Name:     "dns-udp",
		Port:     53,
		HostPort: 53,
		Protocol: "UDP",
	}
	tcpPort := deployer.Port{
		Name:     "dns-tcp",
		Port:     53,
		HostPort: 53,
		Protocol: "TCP",
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: b.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "bind",
					Name:      "bind",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:     b.Context,
			Namespace:   "bind",
			Name:        "bind",
			HostNetwork: true,
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "bind",
					Image: image,
					Requirement: &build.Bind{
						Image:     image,
						GitBranch: docker.GitBranch(version),
					},
					Ports: []deployer.Port{tcpPort, udpPort},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "200m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "25Mi",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "bind",
							Path: "/etc/bind",
						},
						{
							Name: "bind",
							Path: "/var/lib/bind",
						},
					},
					LivenessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: tcpPort.Port,
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					ReadinessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: tcpPort.Port,
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "bind",
					Host: k8s.PodVolumeHost{
						Path: "/data/bind",
					},
				},
			},
		},
	}
}

func (b *Bind) Applier() (world.Applier, error) {
	return nil, nil
}
